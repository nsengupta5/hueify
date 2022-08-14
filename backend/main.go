package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	Generator "hueify/generator"
	Queue "hueify/queue"

	"image"
	"log"
	"net/http"
	"strconv"
	"strings"
)

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

type ArtistRelations map[string]map[string]string

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
	Generator.DrawPlayistCover(0, 40, 150, 30, 200, 100)

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
		api.GET("/get-album/:uri", getAlbum)
		api.POST("/retrieve-new-music/", getNewAlbums)
		api.GET("/related/:uri", getAllRelatedArtists)
		api.POST("/create-playlist/", createPlaylist)

	}

	err = router.Run("localhost:8080")
	if err != nil {
		return
	}
}

func createPlaylist(c *gin.Context) {
	myError := Generator.DrawPlayistCover(0, 255, 255, 100, 0, 0)
	if myError != nil {
		return
	}

	c.AbortWithStatus(200)
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

func getAlbum(c *gin.Context) {

	uri := c.Param("uri")
	id := strings.Split(uri, ":")[2]

	album, err := client.GetAlbum(spotify.ID(id))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "error retrieve album data",
			"error":   err,
		})
	}

	img, err := loadImage(album.Images[0].URL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to load image",
			"error":   err,
		})
	}

	noCroppingColours, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentNoCropping, img)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process image",
			"error":   err,
		})
	}

	croppingColours, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentDefault, img)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to process image",
			"error":   err,
		})
	}

	//use hashmap

	//for i in range croppingColoursExtra
	for i, c := range croppingColours {
		originalColor, isSimilar, index := rgbDiff(c, noCroppingColours, 50)
		// originalColorExtra, isSimilarExtra, indexExtra := rgbDiff(c, croppingColoursExtra, 50)

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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to get related artists",
			"error":   err,
		})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to json marshal album",
			"error":   err,
		})
	}

	var res AlbumRes
	json.Unmarshal(b, &res)

	//Use Context.JSON() instead in production
	//as indentedJson is CPU intensive
	c.IndentedJSON(http.StatusOK, res)

}

func getAllRelatedArtists(c *gin.Context) {
	relatedStruct := make(ArtistRelations)
	visitedArtists := make(map[string]bool)
	queue := Queue.New()

	id := spotify.ID(strings.Split(c.Param("uri"), ":")[2])

	artist, _ := client.GetArtist(id)
	nameAndID := []string{artist.Name, artist.ID.String()}
	depth := 0
	queue = Queue.Enqueue(queue, []string{strings.Join(nameAndID, "|"), strconv.FormatInt(int64(depth), 10)})

	//start := time.Now(); time.Since(start) < 2*time.Second;
	for depth < 4 && len(queue) != 0 {
		var artistName string
		var artistID string
		queue, artistName, artistID, depth = Queue.Dequeue(queue)

		relatedArtistNames, _, err := getRelatedArtists(relatedStruct, spotify.ID(artistID), 10)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to get related artists",
				"error":   err,
			})
		}

		relatedStruct[artistName] = relatedArtistNames

		depth = depth + 1
		for relatedArtistName, relatedArtistID := range relatedArtistNames {
			if _, has := visitedArtists[relatedArtistName]; !has {
				visitedArtists[relatedArtistName] = true
				completeString := []string{relatedArtistName, relatedArtistID}
				queue = Queue.Enqueue(queue, []string{strings.Join(completeString, "|"), strconv.FormatInt(int64(depth), 10)})
			}
		}
	}

	c.IndentedJSON(http.StatusOK, relatedStruct)

}

func getRelatedArtists(relatedStruct ArtistRelations, id spotify.ID, count int) (map[string]string, int, error) {
	relatedArtistsInfo := make(map[string]string)

	fmt.Println("id", id)
	relatedArtists, err := client.GetRelatedArtists(id)
	if err != nil {
		return nil, 0, err
	}

	for i, artist := range relatedArtists {
		if i == count {
			break
		}
		if _, has := relatedStruct[artist.Name]; !has {
			relatedArtistsInfo[artist.Name] = artist.ID.String()
		}
	}

	return relatedArtistsInfo, len(relatedArtistsInfo), nil
}

func getNewAlbums(c *gin.Context) {
	return
}

func rgbDiff(
	color prominentcolor.ColorItem,
	originalColorScheme []prominentcolor.ColorItem,
	difference uint32) (colors prominentcolor.ColorItem, matches bool, index int) {

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

func similarOccurences(occurences uint32, comparingToOccurences uint32) bool {

	percentageDiff := float64(comparingToOccurences) * 0.3
	max := float64(comparingToOccurences) + percentageDiff
	min := float64(comparingToOccurences) - percentageDiff

	if float64(occurences) <= max && float64(occurences) >= min {
		return true
	}

	return false
}

func remove[T any, K Index](s []T, i K) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
