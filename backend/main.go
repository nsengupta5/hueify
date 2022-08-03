package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Color interface {
	prominentcolor.ColorItem
}

type Index interface {
	uint16 | int
}

type rgbRanges struct {
	redMax, redMin     uint32
	greenMax, greenMin uint32
	blueMax, blueMin   uint32
}

type AlbumRes struct {
	Artist             string                     `json:"artist"`
	AlbumImg           string                     `json:"album_image"`
	AlbumName          string                     `json:"album_name"`
	ImageColors        []prominentcolor.ColorItem `json:"image_colors"`
	RelatedArtists     []string                   `json:"related_artists"`
	RelatedArtistsURIs []string                   `json:"related_artists_uri"`
}

type RecommendedAlbum struct {
	Type    string   `json:"type"`
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Artists []string `json:"artists"`
	Image   string   `json:"image"`
}

var authConfig *clientcredentials.Config
var accessToken *oauth2.Token
var client spotify.Client

func main() {
	authConfig = &clientcredentials.Config{
		ClientID:     "f1cfc1de2b5c4b419b2c8e5c50ccd4e1",
		ClientSecret: "f1e1873798744ca29a7e208f9cafb73c",
		TokenURL:     spotify.TokenURL,
	}

	var err error
	accessToken, err = authConfig.Token(context.Background())
	if err != nil {
		log.Fatalf("error retrieve access token: %v", err)
	}

	client = spotify.Authenticator{}.NewClient(accessToken)

	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Serve frontend static files
	router.Use(static.Serve("/", static.LocalFile("../frontend/build", true)))

	// Routes
	api := router.Group("/api")
	{
		api.GET("/", root)
		api.GET("/artist/:uri", getArtist)
		api.GET("/get-album/:uri", getAlbum)
		api.POST("/retrieve-new-music", getNewAlbums)

	}

	err = router.Run("localhost:8080")
	if err != nil {
		return
	}
}

func loadImage(fileInput string) (image.Image, error) {
	res, err := http.Get(fileInput)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	img, _, err := image.Decode(res.Body)
	return img, err
}

func getArtist(c *gin.Context) {
	uri := c.Param("uri")
	fmt.Println(uri)
	id := strings.Split(uri, ":")[2]

	artistID := spotify.ID(id)
	artist, err := client.GetArtist(artistID)
	if err != nil {
		log.Fatalf("error retrieve artist data: %v", err)
	}

	c.IndentedJSON(http.StatusOK, artist)
}

func getAlbum(c *gin.Context) {
	uri := c.Param("uri")
	id := strings.Split(uri, ":")[2]

	album, err := client.GetAlbum(spotify.ID(id))
	if err != nil {
		log.Fatalf("error retrieve album data: %v", err)
	}

	img, err := loadImage(album.Images[0].URL)
	if err != nil {
		log.Fatal("Failed to load image", err)
	}

	noCroppingColours, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentNoCropping, img)
	if err != nil {
		log.Fatal("Failed to process image", err)
	}

	croppingColours, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentDefault, img)
	if err != nil {
		log.Fatal("Failed to process image", err)
	}

	for i, c := range croppingColours {
		originalColor, isSimilar, index := rgbDiff(c, noCroppingColours, 50)
		if isSimilar {
			if originalColor.Cnt > c.Cnt {
				//delete the c from croppingColours
				croppingColours = remove(croppingColours, i)
			} else {
				//delete the originalColor from noCroppingColours
				noCroppingColours = remove(noCroppingColours, index)
			}
		}
	}

	jointColours := append(noCroppingColours, croppingColours...)

	relatedArtists, err := client.GetRelatedArtists(album.Artists[0].ID)
	if err != nil {
		log.Fatal("Failed to get related artists", err)
	}

	relatedArtistsURIs := make([]string, 0)
	relatedArtistsNames := make([]string, 0)

	for _, a := range relatedArtists {
		relatedArtistsURIs = append(relatedArtistsURIs, string(a.URI))
		relatedArtistsNames = append(relatedArtistsNames, a.Name)
	}

	albumRes := &AlbumRes{
		Artist:             album.Artists[0].Name,
		AlbumImg:           album.Images[0].URL,
		AlbumName:          album.Name,
		ImageColors:        jointColours,
		RelatedArtists:     relatedArtistsNames,
		RelatedArtistsURIs: relatedArtistsURIs,
	}

	b, err := json.Marshal(albumRes)
	if err != nil {
		log.Fatal("Failed to json marshal album", err)
	}

	var res AlbumRes
	json.Unmarshal(b, &res)

	//Use Context.JSON() instead in production
	//as indentedJson is CPU intensive
	c.IndentedJSON(http.StatusOK, res)

}

func rgbDiff(
	color prominentcolor.ColorItem,
	originalColorScheme []prominentcolor.ColorItem,
	difference uint32) (prominentcolor.ColorItem, bool, int) {

	myRGBRanges := rgbRanges{
		redMax:   color.Color.R + difference,
		redMin:   color.Color.R - difference,
		greenMax: color.Color.G + difference,
		greenMin: color.Color.G - difference,
		blueMax:  color.Color.B + difference,
		blueMin:  color.Color.B - difference,
	}

	for i, c := range originalColorScheme {
		if similarColor(c, myRGBRanges) {
			return c, true, i
		}
	}

	return prominentcolor.ColorItem{}, false, 1
}

func similarColor(color prominentcolor.ColorItem, rgb rgbRanges) bool {
	isSimilarRed := color.Color.R <= rgb.redMax && color.Color.R >= rgb.redMin
	isSimilarGreen := color.Color.G <= rgb.greenMax && color.Color.G >= rgb.greenMin
	isSimilarBlue := color.Color.B <= rgb.blueMax && color.Color.B >= rgb.blueMin

	if isSimilarRed && isSimilarGreen && isSimilarBlue {
		return true
	}

	return false
}

func remove[T any, K Index](s []T, i K) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func root(c *gin.Context) {
	return
}

func getNewAlbums(c *gin.Context) {
	return
}
