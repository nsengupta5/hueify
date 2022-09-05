package main

import (
	"context"
	"encoding/json"
	"fmt"
	Generator "hueify/generator"
	Queue "hueify/queue"
	"image"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Options spotify.Options

type Index interface {
	uint16 | int
}

type rgbRanges struct {
	redMax, redMin     uint32
	greenMax, greenMin uint32
	blueMax, blueMin   uint32
}

type RelatedArtistInfo struct {
	Id         string
	Popularity int32
}

type AlbumRes struct {
	Artist             string                     `json:"artist"`
	ArtistId           spotify.ID                 `json:"artist_id"`
	AlbumImg           string                     `json:"album_image"`
	AlbumName          string                     `json:"album_name"`
	AlbumId            spotify.ID                 `json:"album_id"`
	ImageColors        []prominentcolor.ColorItem `json:"image_colors"`
	RelatedArtists     []string                   `json:"related_artists"`
	RelatedArtistsURIs []string                   `json:"related_artists_uri"`
}

type Info struct {
	Related    map[string]Info
	Popularity int
}

type ArtistRelations map[string]Info

type RecommendedAlbumReq struct {
	ColorScheme []prominentcolor.ColorItem `json:"colorScheme"`
	URI         string                     `json:"uri"`
}

type RecommendedAlbum struct {
	Type      string `json:"type"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Artists   string `json:"artists"`
	Image     string `json:"image"`
	EndStream bool   `json:"endStream"`
}

var authConfig *clientcredentials.Config
var accessToken *oauth2.Token
var client spotify.Client

func main() {
	//Generator.DrawPlayistCover(255, 40, 150, 30, 200, 100)

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
		api.GET("/get-album/:uri", getAlbumReq)
		api.GET("/recommended/:albumId/:artistId", getNewAlbums)
		api.GET("/related/:uri", getAllRelatedArtistsReq)
		api.POST("/create-playlist/", createPlaylist)

	}

	err = router.Run("localhost:8080")
	if err != nil {
		return
	}
}

func getAllRelatedArtistsReq(c *gin.Context) {
	id := spotify.ID(strings.Split(c.Param("uri"), ":")[2])

	artists, err := getAllRelatedArtists(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, err)
	}

	c.IndentedJSON(http.StatusOK, artists)
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

func getAlbumReq(c *gin.Context) {
	uri := c.Param("uri")

	album, err := getAlbum(uri, true)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Failed to retrieve album info",
			"error":   err,
		})
	}

	b, err := json.Marshal(&album)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to json marshal album",
			"error":   err,
		})
	}

	var res AlbumRes
	json.Unmarshal(b, &res)

	//Use c.JSON() instead in production
	//as indentedJson is CPU intensive
	c.IndentedJSON(http.StatusOK, res)
}

func getAlbum(identifier string, isURI bool) (AlbumRes, error) {

	var id string
	if isURI {
		id = strings.Split(identifier, ":")[2]
	} else {
		id = identifier
	}

	album, err := client.GetAlbum(spotify.ID(id))
	if err != nil {
		return AlbumRes{}, err
		//c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		//	"message": "error retrieve album data",
		//	"error":   err,
		//})
	}

	img, err := loadImage(album.Images[0].URL)
	if err != nil {
		return AlbumRes{}, err
		/*		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to load image",
				"error":   err,
			})*/
	}

	noCroppingColours, err := prominentcolor.KmeansWithAll(6, img, prominentcolor.ArgumentNoCropping, 640, nil)
	if err != nil {
		return AlbumRes{}, err
		//c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		//	"message": "Failed to process image",
		//	"error":   err,
		//})
	}

	relatedArtists, err := client.GetRelatedArtists(album.Artists[0].ID)
	if err != nil {
		return AlbumRes{}, err
		//c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		//	"message": "Failed to get related artists",
		//	"error":   err,
		//})
	}

	//noCroppingColours = removeSimilarColor(noCroppingColours)

	relatedArtistsURIs := make([]string, 0)
	relatedArtistsNames := make([]string, 0)

	for _, a := range relatedArtists {
		relatedArtistsURIs = append(relatedArtistsURIs, string(a.URI))
		relatedArtistsNames = append(relatedArtistsNames, a.Name)
	}

	albumRes := AlbumRes{
		Artist:             album.Artists[0].Name,
		ArtistId:           album.Artists[0].ID,
		AlbumImg:           album.Images[0].URL,
		AlbumName:          album.Name,
		AlbumId:            album.ID,
		ImageColors:        noCroppingColours,
		RelatedArtists:     relatedArtistsNames,
		RelatedArtistsURIs: relatedArtistsURIs,
	}

	return albumRes, nil

}

func getAllRelatedArtists(id spotify.ID) ([]RelatedArtistInfo, error) {
	relatedStruct := make(ArtistRelations)
	visitedArtists := make(map[string]bool)
	queue := Queue.New()

	artist, err := client.GetArtist(id)
	if err != nil {
		return nil, err
	}

	idAndPopularity := []string{artist.ID.String(), strconv.Itoa(artist.Popularity)}
	depth := 0
	queue = Queue.Enqueue(queue, []string{strings.Join(idAndPopularity, "|"), strconv.FormatInt(int64(depth), 10)})

	for depth < 4 && len(queue) != 0 {
		var artistPopularity string
		var artistID string
		queue, artistID, artistPopularity, depth = Queue.Dequeue(queue)

		relatedArtistNames, _, err := getRelatedArtists(relatedStruct, spotify.ID(artistID), 10)
		if err != nil {
			return nil, err
		}

		artistPopularityInt, _ := strconv.Atoi(artistPopularity)

		relatedStruct[artistID] = Info{relatedArtistNames, artistPopularityInt}

		depth = depth + 1
		for relatedArtistID, relatedArtistInfo := range relatedArtistNames {
			if _, has := visitedArtists[relatedArtistID]; !has {
				visitedArtists[relatedArtistID] = true
				relatedArtistPopularityStr := strconv.Itoa(relatedArtistInfo.Popularity)
				completeString := []string{relatedArtistID, relatedArtistPopularityStr}
				queue = Queue.Enqueue(queue, []string{strings.Join(completeString, "|"), strconv.FormatInt(int64(depth), 10)})
			}
		}
	}

	//initialize slice
	listOfRelated := make([]RelatedArtistInfo, 0)

	//convert map to slice. this makes the recommended albums process easier when dealing with slice (traversing is easy)
	for key, value := range relatedStruct {
		listOfRelated = append(listOfRelated, RelatedArtistInfo{key, int32(value.Popularity)})
		for childKey, childValue := range value.Related {
			listOfRelated = append(listOfRelated, RelatedArtistInfo{childKey, int32(childValue.Popularity)})
		}
	}

	//sort based on popularity
	sort.Slice(listOfRelated, func(d, e int) bool {
		return listOfRelated[d].Popularity > listOfRelated[e].Popularity
	})

	return listOfRelated, nil
}

func getRelatedArtists(relatedStruct ArtistRelations, id spotify.ID, count int) (related map[string]Info, length int, err error) {
	relatedArtistsInfo := make(map[string]Info)

	fmt.Println("id", id)
	relatedArtists, err := client.GetRelatedArtists(id)
	if err != nil {
		return nil, 0, err
	}

	for i, artist := range relatedArtists {
		if i == count {
			break
		}
		if _, has := relatedStruct[artist.ID.String()]; !has {
			relatedArtistsInfo[artist.ID.String()] = Info{map[string]Info{}, artist.Popularity}
		}
	}

	return relatedArtistsInfo, len(relatedArtistsInfo), nil
}

func getNewAlbums(c *gin.Context) {

	c.Header("Content-Type", "text/event-stream")

	//var request RecommendedAlbumReq
	//if err := c.ShouldBindJSON(&request); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//	return
	//}

	albumId := c.Param("albumId")
	artistId := c.Param("artistId")
	album, err := getAlbum(albumId, false)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "error retrieve album data",
			"error":   err,
		})
	}

	relatedArtists, err := getAllRelatedArtists(spotify.ID(artistId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get related artists",
			"error":   err,
		})
	}

	//create channel which stores the recommended albums
	recommended := make(chan RecommendedAlbum, 6)

	//originalColorScheme := request.ColorScheme
	originalColorScheme := album.ImageColors

	//split related artists list into 3 slices
	//for len(recommended) != cap(recommended) {
	//	go func() {
	//		recommended, err = searchAlbums(relatedArtists[:len(relatedArtists)/2], originalColorScheme, recommended)
	//		if err != nil {
	//
	//		}
	//	}()
	//	go func() {
	//		recommended, err = searchAlbums(relatedArtists[len(relatedArtists)/2:], originalColorScheme, recommended)
	//		if err != nil {
	//
	//		}
	//	}()
	//}
	recommended, err = searchAlbums(relatedArtists[0:50], originalColorScheme, recommended, c)

	fmt.Println(&recommended)
	c.AbortWithStatus(http.StatusOK)
}

func searchAlbums(
	relatedArtistsSlice []RelatedArtistInfo,
	originalColorScheme []prominentcolor.ColorItem,
	ch chan RecommendedAlbum,
	c *gin.Context) (chan RecommendedAlbum, error) {

	visitedAlbums := map[string]bool{}

	//for each artist
	for i, artist := range relatedArtistsSlice {
		println("looking at artist:" + strconv.Itoa(i))
		//get albums
		//id := strings.Split(artist.Id, ":")[2]
		albums, err := client.GetArtistAlbums(spotify.ID(artist.Id))
		if err != nil {
			return ch, err
		}

		for j, album := range albums.Albums {

			println("looking at album:" + strconv.Itoa(j))

			if visitedAlbums[album.ID.String()] {
				continue
			}

			//get color scheme of album
			img, err := loadImage(album.Images[0].URL)
			if err != nil {
				return ch, err
			}

			colors, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentNoCropping, img)
			if err != nil {
				return ch, err
			}

			//compare colors to original color scheme
			similarColorsCount := compareArtwork(originalColorScheme, colors)
			ratio := float32(similarColorsCount / len(colors))
			//if 50% match then write to channel
			if visited := visitedAlbums[album.ID.String()]; ratio >= 0.5 && !visited {
				fmt.Println("color match:")

				endStream := len(ch) == cap(ch)

				albumToReturn := RecommendedAlbum{
					Type:      album.AlbumType,
					Id:        album.ID.String(),
					Name:      album.Name,
					Artists:   album.Artists[0].Name,
					Image:     album.Images[0].URL,
					EndStream: endStream,
				}

				visitedAlbums[album.ID.String()] = true

				print(albumToReturn.Name)

				ch <- albumToReturn

				c.Stream(func(w io.Writer) bool {
					if msg, ok := <-ch; ok {
						c.SSEvent("message", msg)
						return false
					}
					return true
				})

				println("got an album")
			} else {
				visitedAlbums[album.ID.String()] = true
			}
		}
	}

	return ch, nil
}

func compareArtwork(original []prominentcolor.ColorItem, current []prominentcolor.ColorItem) int {
	count := 0

	for _, currentColor := range current {
		color, matches, index := rgbDiff(currentColor, original, 25)
		if matches {
			if similarOccurences(uint32(color.Cnt), uint32(original[index].Cnt)) {
				count = count + 1
			}
		}
	}

	return count
}

func Combinations(iterable []prominentcolor.ColorItem, r int) (rt [][]prominentcolor.ColorItem) {
	pool := iterable
	n := len(pool)

	if r > n {
		return
	}

	indices := make([]int, r)
	for i := range indices {
		indices[i] = i
	}

	result := make([]prominentcolor.ColorItem, r)
	for i, el := range indices {
		result[i] = pool[el]
	}
	s2 := make([]prominentcolor.ColorItem, r)
	copy(s2, result)
	rt = append(rt, s2)

	for {
		i := r - 1
		for ; i >= 0 && indices[i] == i+n-r; i -= 1 {
		}

		if i < 0 {
			return
		}

		indices[i] += 1
		for j := i + 1; j < r; j += 1 {
			indices[j] = indices[j-1] + 1
		}

		for ; i < len(indices); i += 1 {
			result[i] = pool[indices[i]]
		}
		s2 = make([]prominentcolor.ColorItem, r)
		copy(s2, result)
		rt = append(rt, s2)
	}
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

func removeSimilarColor(
	originalColorScheme []prominentcolor.ColorItem) []prominentcolor.ColorItem {
	newColorScheme := originalColorScheme[:6]
	return newColorScheme
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
