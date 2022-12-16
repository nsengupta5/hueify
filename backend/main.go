package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
	Generator "hueify/generator"
	HttpError "hueify/http-errors"
	Queue "hueify/queue"
	Structs "hueify/structs"
	"image"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/hisamafahri/coco"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var authConfig *clientcredentials.Config
var accessToken *oauth2.Token
var client spotify.Client

var ctx context.Context
var sa option.ClientOption
var app *firebase.App
var firestoreClient *firestore.Client

var firebaseErr error
var clientErr error

func main() {
	var err error

	//set up Firestore for artist metadata (related artists)
	ctx = context.Background()
	sa = option.WithCredentialsFile("./firebase.json")
	app, firebaseErr = firebase.NewApp(ctx, nil, sa)
	if firebaseErr != nil {
		log.Fatalln(firebaseErr)
	}

	firestoreClient, clientErr = app.Firestore(ctx)
	if clientErr != nil {
		log.Fatalln(clientErr)
	}

	defer func(client *firestore.Client) {
		err := client.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(firestoreClient)

	authConfig = &clientcredentials.Config{
		ClientID:     "f1cfc1de2b5c4b419b2c8e5c50ccd4e1",
		ClientSecret: "f1e1873798744ca29a7e208f9cafb73c",
		TokenURL:     spotify.TokenURL,
	}

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

func getColors(img image.Image) ([]prominentcolor.ColorItem, error) {
	return prominentcolor.KmeansWithAll(6, img, prominentcolor.ArgumentNoCropping, 640, nil)
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
		HttpError.AlbumInfoFailure(c, err)
	}

	b, err := json.Marshal(&album)
	if err != nil {
		HttpError.JSONSerializeFailure(c, err)
	}

	var res Structs.AlbumRes
	err = json.Unmarshal(b, &res)
	if err != nil {
		HttpError.JSONDeserializeFailure(c, err)
	}
	//Use c.JSON() instead in production
	//as indentedJson is CPU intensive
	c.IndentedJSON(http.StatusOK, res)
}

func getAlbum(identifier string, isURI bool) (Structs.AlbumRes, error) {

	var id string
	var newReq bool

	if isURI {
		id = strings.Split(identifier, ":")[2]
	} else {
		id = identifier
	}

	album, err := client.GetAlbum(spotify.ID(id))
	if err != nil {
		return Structs.AlbumRes{}, err
	}

	artistName := album.SimpleAlbum.Artists[0].Name
	_, err = firestoreClient.Doc("artists/" + artistName).Get(ctx)
	if err != nil {
		newReq = true
	} else {
		newReq = false
	}

	img, err := loadImage(album.Images[0].URL)
	if err != nil {
		return Structs.AlbumRes{}, err
	}

	noCroppingColours, err := getColors(img)
	if err != nil {
		return Structs.AlbumRes{}, err
	}

	relatedArtists, err := client.GetRelatedArtists(album.Artists[0].ID)
	if err != nil {
		return Structs.AlbumRes{}, err
	}

	relatedArtistsURIs := make([]string, 0)
	relatedArtistsNames := make([]string, 0)

	for _, a := range relatedArtists {
		relatedArtistsURIs = append(relatedArtistsURIs, string(a.URI))
		relatedArtistsNames = append(relatedArtistsNames, a.Name)
	}

	albumRes := Structs.AlbumRes{
		Artist:             album.Artists[0].Name,
		ArtistId:           album.Artists[0].ID,
		AlbumImg:           album.Images[0].URL,
		AlbumName:          album.Name,
		AlbumId:            album.ID,
		ImageColors:        noCroppingColours,
		RelatedArtists:     relatedArtistsNames,
		RelatedArtistsURIs: relatedArtistsURIs,
		NewReq:             newReq,
	}

	return albumRes, nil

}

func getAllRelatedArtists(id spotify.ID) ([]Structs.RelatedArtistInfo, error) {
	relatedStruct := make(Structs.ArtistRelations)
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

		relatedStruct[artistID] = Structs.Info{Related: relatedArtistNames, Popularity: artistPopularityInt}

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
	listOfRelated := make([]Structs.RelatedArtistInfo, 0)

	//convert map to slice. this makes the recommended albums process easier when dealing with slice (traversing is easy)
	for key, value := range relatedStruct {
		listOfRelated = append(listOfRelated, Structs.RelatedArtistInfo{Id: key, Popularity: int32(value.Popularity)})
		for childKey, childValue := range value.Related {
			listOfRelated = append(listOfRelated, Structs.RelatedArtistInfo{Id: childKey, Popularity: int32(childValue.Popularity)})
		}
	}

	//sort based on popularity
	sort.Slice(listOfRelated, func(d, e int) bool {
		return listOfRelated[d].Popularity > listOfRelated[e].Popularity
	})

	//add to firestore
	//_, err = firestoreClient.Collection("artists").Doc(artist.Name).Set(ctx, listOfRelated)

	//UNCOMMENT LATER
	//_, err = firestoreClient.Collection("artists").Doc(artist.Name).Create(ctx, map[string]interface{}{
	//	"related": listOfRelated,
	//})
	//if err != nil {
	//	log.Fatalf("Failed adding artist to firestore: %v", err)
	//}

	return listOfRelated, nil
}

func getRelatedArtists(relatedStruct Structs.ArtistRelations, id spotify.ID, count int) (related map[string]Structs.Info, length int, err error) {
	relatedArtistsInfo := make(map[string]Structs.Info)

	//fmt.Println("id", id)
	relatedArtists, err := client.GetRelatedArtists(id)
	if err != nil {
		return nil, 0, err
	}

	for i, artist := range relatedArtists {
		if i == count {
			break
		}
		if _, has := relatedStruct[artist.ID.String()]; !has {
			relatedArtistsInfo[artist.ID.String()] = Structs.Info{Related: map[string]Structs.Info{}, Popularity: artist.Popularity}
		}
	}

	return relatedArtistsInfo, len(relatedArtistsInfo), nil
}

func getNewAlbums(c *gin.Context) {

	c.Header("Content-Type", "text/event-stream")

	//var request Structs.RecommendedAlbumReq
	//if err := c.ShouldBindJSON(&request); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//	return
	//}

	albumId := c.Param("albumId")
	artistId := c.Param("artistId")
	album, err := getAlbum(albumId, false)

	if err != nil {
		HttpError.AlbumInfoFailure(c, err)
	}

	var relatedArtists []Structs.RelatedArtistInfo
	//var ok bool

	//if album.NewReq {
	relatedArtists, err = getAllRelatedArtists(spotify.ID(artistId))
	if err != nil {
		HttpError.GetRelatedArtistsFailure(c, err)
	}
	//} else {
	//	artistDoc, err := firestoreClient.Collection("artists").Doc(album.Artist).Get(ctx)
	//
	//	artistDoc.DataTo(&relatedArtists);
	//
	//	if err != nil {
	//		HttpError.GetRelatedArtistsFailure(c, err)
	//	}
	//	relatedArtistsInterface := artistDoc.Data()["related"]
	//
	//	switch reflect.TypeOf(relatedArtistsInterface).Kind() {
	//	case reflect.Slice:
	//		s := reflect.ValueOf(relatedArtistsInterface)
	//
	//		for i := 0; i < s.Len(); i++ {
	//		}
	//	}
	//
	//	//relatedArtistsInterface := coll["related"]
	//
	//	for _, relatedArtist := range relatedArtistsInterface.([]RelatedArtistInfo) {
	//		relatedArtists = append(relatedArtists, relatedArtist)
	//	}
	//
	//	//relatedArtists, ok := relatedArtistsInterface.([]RelatedArtistInfo)
	//	//if !ok {
	//	//	println(ok)
	//	//}
	//
	//	fmt.Printf("Document: %#v\\n", relatedArtists)
	//
	//	if err != nil {
	//		HttpError.GetCachedRelatedArtistsFailure(c, err)
	//	}
	//}

	//create channel which stores the recommended albums
	recommended := make(chan Structs.RecommendedAlbum, 6)

	//originalColorScheme := request.ColorScheme
	originalColorScheme := album.ImageColors

	bound1 := float64(len(relatedArtists)) * 0.1
	bound2 := float64(len(relatedArtists)) * 0.2
	bound3 := float64(len(relatedArtists)) * 0.3
	bound4 := float64(len(relatedArtists)) * 0.4
	bound5 := float64(len(relatedArtists)) * 0.5
	bound6 := float64(len(relatedArtists)) * 0.6
	bound7 := float64(len(relatedArtists)) * 0.7
	bound8 := float64(len(relatedArtists)) * 0.8
	bound9 := float64(len(relatedArtists)) * 0.9
	bounds := [10]int{0, int(bound1), int(bound2), int(bound3), int(bound4), int(bound5), int(bound6), int(bound7),
		int(bound8), int(bound9)}

	go func() {
		err = searchAlbums(relatedArtists[:bounds[0]], originalColorScheme, recommended)
		if err != nil {
			HttpError.GetRecommendedAlbumsFailure(c, err)
		}
	}()

	for i := 1; i < len(bounds)-1; i++ {
		go func(_i int) {
			err = searchAlbums(relatedArtists[bounds[_i]:bounds[_i+1]], originalColorScheme, recommended)
			if err != nil {
				HttpError.GetRecommendedAlbumsFailure(c, err)
			}
		}(i)
	}

	go func() {
		err = searchAlbums(relatedArtists[bounds[9]:], originalColorScheme, recommended)
		if err != nil {
			HttpError.GetRecommendedAlbumsFailure(c, err)
		}
	}()

	//spin lock implementation
	done := false
	for !done {
		if len(recommended) == cap(recommended) {
			done = true
		} else {
			//if new album added to channel then stream response
			c.Stream(func(w io.Writer) bool {
				if contents, ok := <-recommended; ok {
					c.SSEvent("message", contents)
					return false
				}
				return true
			})
		}
	}

	fmt.Println(&recommended)
	c.AbortWithStatus(http.StatusOK)
}

func searchAlbums(
	relatedArtistsSlice []Structs.RelatedArtistInfo,
	originalColorScheme []prominentcolor.ColorItem,
	ch chan Structs.RecommendedAlbum) error {

	var routineClient spotify.Client
	routineClient = spotify.Authenticator{}.NewClient(accessToken)

	visitedAlbums := map[string]bool{}

	//for each artist
	for i, artist := range relatedArtistsSlice {
		println("looking at artist:" + strconv.Itoa(i))
		//get albums
		albums, err := routineClient.GetArtistAlbums(spotify.ID(artist.Id))
		if err != nil {
			return err
		}

		for j, album := range albums.Albums {

			println("looking at album:" + strconv.Itoa(j))

			if visitedAlbums[album.ID.String()] {
				continue
			}

			//get color scheme of album
			img, err := loadImage(album.Images[0].URL)
			if err != nil {
				return err
			}

			colors, err := getColors(img)
			if err != nil {
				return err
			}

			//compare colors to original color scheme
			artworkIsSimilar := compareArtworkNew(originalColorScheme, colors)

			//if 50% match then write to channel
			if visited := visitedAlbums[album.ID.String()]; artworkIsSimilar && !visited {
				fmt.Println("color match:")

				endStream := len(ch) == cap(ch)

				albumToReturn := Structs.RecommendedAlbum{
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

				println("got an album")
			} else {
				visitedAlbums[album.ID.String()] = true
			}
		}
	}

	return nil
}

func compareArtworkNew(original []prominentcolor.ColorItem, current []prominentcolor.ColorItem) bool {
	palette_len := len(original)

	difference := float64(25)
	for i := 0; i < palette_len/2; i++ {
		if betterSimilarColor(original[i], current[i]) <= difference {
			if i == 0 {
				difference += 2
			} else {
				difference += 5
			}
		} else {
			return false
		}
	}
	return true
}

func add(colorVal uint32, valToChange uint32) uint32 {
	changed := int(colorVal) - int(valToChange)

	if changed > 255 {
		return 255
	}

	return colorVal + valToChange
}

func deduct(colorVal uint32, valToChange uint32) uint32 {
	changed := int(colorVal) - int(valToChange)

	if changed < 0 {
		return 0
	}

	return colorVal - valToChange
}

func rgbDiff(
	color prominentcolor.ColorItem,
	originalColorScheme []prominentcolor.ColorItem,
	difference uint32) (colors prominentcolor.ColorItem, matches bool, index int) {

	myRGBRanges := Structs.RGBRanges{
		RedMax:   add(color.Color.R, difference),
		RedMin:   deduct(color.Color.R, difference),
		GreenMax: add(color.Color.G, difference),
		GreenMin: deduct(color.Color.G, difference),
		BlueMax:  add(color.Color.B, difference),
		BlueMin:  deduct(color.Color.B, difference),
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

func betterSimilarColor(color1 prominentcolor.ColorItem, color2 prominentcolor.ColorItem) float64 {
	color1_lab := coco.Rgb2Lab(float64(color1.Color.R), float64(color1.Color.G), float64(color1.Color.B))
	color2_lab := coco.Rgb2Lab(float64(color2.Color.R), float64(color2.Color.G), float64(color2.Color.B))

	color1_L := color1_lab[0]
	color2_L := color2_lab[0]
	color1_a := color1_lab[1]
	color2_a := color2_lab[1]
	color1_b := color1_lab[2]
	color2_b := color2_lab[2]

	// For graphic arts
	kl := 1
	k1 := 0.045
	k2 := 0.015
	kc := float64(1)
	kh := float64(1)

	delta_L := color1_L - color2_L
	c1 := math.Sqrt(math.Pow(color1_a, 2) + math.Pow(color1_b, 2))
	c2 := math.Sqrt(math.Pow(color2_a, 2) + math.Pow(color2_b, 2))
	delta_c := c1 - c2
	delta_a := color1_a - color2_a
	delta_b := color1_b - color2_b
	delta_h := math.Sqrt(math.Pow(delta_a, 2) + math.Pow(delta_b, 2) - math.Pow(delta_c, 2))
	sl := 1
	sc := 1 + k1*c1
	sh := 1 + k2*c1

	delta_e := math.Sqrt(math.Pow((delta_L/float64(kl*sl)), 2) + math.Pow((delta_c/float64(kc*sc)), 2) + math.Pow((delta_h/kh*sh), 2))

	/* Delta E Values:
	<= 1 --> not perceptible
	1-2 --> perceptible close observation
	2-10 --> perceptible at glance
	11-49 --> colors more similar than opposite
	100 --> colors exactly opposite
	*/
	return delta_e
}

func similarColor(color prominentcolor.ColorItem, rgb Structs.RGBRanges) bool {
	isSimilarRed := color.Color.R <= rgb.RedMax && color.Color.R >= rgb.RedMin
	isSimilarGreen := color.Color.G <= rgb.GreenMax && color.Color.G >= rgb.GreenMin
	isSimilarBlue := color.Color.B <= rgb.BlueMax && color.Color.B >= rgb.BlueMin

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

func remove[T any, K Structs.Index](s []T, i K) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
