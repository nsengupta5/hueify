package main

import (
	"context"
	"encoding/json"
	"fmt"
	Generator "hueify/generator"
	HttpError "hueify/http-errors"
	Queue "hueify/queue"
	Structs "hueify/structs"
	"image"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = string(8080)
	}

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

	err = router.Run(":" + port)
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

	if strings.Contains(uri, "track") {
		id := strings.Split(uri, ":")[2]
		track, err := client.GetTrack(spotify.ID(id))
		if err != nil {
			return
		}
		uri = string(track.Album.URI)
	}


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
	return
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
		err = searchAlbums(relatedArtists[:bounds[0]], album, recommended)
		if err != nil {
			HttpError.GetRecommendedAlbumsFailure(c, err)
		}
	}()

	for i := 1; i < len(bounds)-1; i++ {
		go func(_i int) {
			err = searchAlbums(relatedArtists[bounds[_i]:bounds[_i+1]], album, recommended)
			if err != nil {
				HttpError.GetRecommendedAlbumsFailure(c, err)
			}
		}(i)
	}

	go func() {
		err = searchAlbums(relatedArtists[bounds[9]:], album, recommended)
		if err != nil {
			HttpError.GetRecommendedAlbumsFailure(c, err)
		}
	}()

	//spin lock implementation
	done := false
	var latestAlbumReturned Structs.RecommendedAlbum
	latestAlbumReturned = Structs.RecommendedAlbum{}
	for !done {
		if len(recommended) == cap(recommended) {
			done = true
		} else {
			//if new album added to channel then stream response
			c.Stream(func(w io.Writer) bool {
				if contents, ok := <-recommended; ok {
					if latestAlbumReturned.Id != contents.Id {
						latestAlbumReturned = contents
						c.SSEvent("message", contents)
						return false
					}
				}
				return true
			})
		}
	}

	c.AbortWithStatus(http.StatusOK)
	return
}

func exactAlbum(album spotify.SimpleAlbum, albumToCompareTo Structs.AlbumRes) bool {
	if album.Artists[0].Name == albumToCompareTo.Artist && strings.Contains(album.Name, albumToCompareTo.AlbumName) {
		return true
	}
	return false
}

func searchAlbums(
	relatedArtistsSlice []Structs.RelatedArtistInfo,
	originalAlbum Structs.AlbumRes,
	ch chan Structs.RecommendedAlbum) error {

	var routineClient spotify.Client
	routineClient = spotify.Authenticator{}.NewClient(accessToken)

	visitedAlbums := map[string]bool{}
	originalColorScheme := originalAlbum.ImageColors

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

			if exactAlbum(album, originalAlbum) {
				continue
			}

			if visitedAlbums[album.ID.String()] {
				continue
			}

			//get color scheme of album

			//debug index out of range error for album.Images[0]
			if len(album.Images) == 0 {
				return nil
			}

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
					Colors:    colors,
					EndStream: endStream,
				}

				visitedAlbums[album.ID.String()] = true

				print(albumToReturn.Name)

				if album.AlbumType != "single" {
					ch <- albumToReturn
				}

				println("got an album")
			} else {
				visitedAlbums[album.ID.String()] = true
			}
		}
	}

	return nil
}

func getMostSaturatedColor(colors []prominentcolor.ColorItem) []prominentcolor.ColorItem {
	mappings := make(map[prominentcolor.ColorItem]float64)
	
	for i := 0; i < len(colors); i++ {
		_,s,_ := RGBToHSL(colors[i])
		mappings[colors[i]] = s
	}

	p := make(Structs.PairList, len(mappings))
	i := 0
	for k,v := range mappings {
		p[i] = Structs.Pair{k, v}
		i++
	}

	sort.Sort(sort.Reverse(p))

	result := make([]prominentcolor.ColorItem, 0)
	for _, k := range p {
		result = append(result, k.Key)
	}

	return result
}

func compareArtworkNew(original []prominentcolor.ColorItem, current []prominentcolor.ColorItem) bool {
	originalLen := len(original)
	currLen := len(current)
	mostImportant := make([]prominentcolor.ColorItem, 0)
	mostSaturatedArr := getMostSaturatedColor(original)
	mostSaturated := mostSaturatedArr[0]

	if (mostSaturated == original[0]) {
		if (mostSaturatedArr[1] == original[1]) {
			mostSaturated = mostSaturatedArr[2]
		} else {
			mostSaturated = mostSaturatedArr[1]
		}
	}

	if (mostSaturated == original[1]) {
		mostSaturated = mostSaturatedArr[1]
	}

	mostImportant = append(mostImportant, original[0])
	mostImportant = append(mostImportant, original[1])
	mostImportant = append(mostImportant, mostSaturated)

	difference := float64(10)
	for i := 0; i < originalLen/2; i++ {
		found := false
		for j := 0; j < currLen/2; j++ {
			if betterSimilarColor(mostImportant[i], current[j]) <= difference {
				difference = 16
				found = true
				break
			}
		}
		if !found {
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

func RGBtoXYZ(color prominentcolor.ColorItem) (float64, float64, float64) {
	var_R := float64(color.Color.R) / 255
	var_G := float64(color.Color.G) / 255
	var_B := float64(color.Color.B) / 255

	if (var_R > 0.04045) {
		var_R = math.Pow((var_R + 0.055) / 1.055, 2.4)
	} else {
		var_R /= 12.92
	}

	if (var_G > 0.04045) {
		var_G = math.Pow((var_G + 0.055) / 1.055, 2.4)
	} else {
		var_G /= 12.92
	}

	if (var_B > 0.04045) {
		var_B = math.Pow((var_B + 0.055) / 1.055, 2.4)
	} else {
		var_B /= 12.92
	}

	var_R *= 100
	var_G *= 100
	var_B *= 100

	x := var_R * 0.4124 + var_G * 0.3576 + var_B * 0.1805
	y := var_R * 0.2126 + var_G * 0.7152 + var_B * 0.0722
	z := var_R * 0.0193 + var_G * 0.1192 + var_B * 0.9505

	return x,y,z
} 

func XYZToLAB(x float64, y float64, z float64) (float64, float64, float64) {
	reference_x := 95.047
	reference_y := 100.0
	reference_z := 108.883

	var_X := x / reference_x
	var_Y := y / reference_y
	var_Z := z / reference_z

	if (var_X > 0.008856) {
		var_X = math.Pow(var_X, float64(1)/3)
	} else {
		var_X = (7.787 * var_X) + (float64(16) / 116)
	}
	
	if (var_Y > 0.008856) {
		var_Y = math.Pow(var_Y, float64(1)/3)
	} else {
		var_Y = (7.787 * var_Y) + (float64(16) / 116)
	}

	if (var_Z > 0.008856) {
		var_Z = math.Pow(var_Z, float64(1)/3)
	} else {
		var_Z = (7.787 * var_Z) + (float64(16) / 116)
	}

	l := (116 * var_Y) - 16
	a := 500 * (var_X - var_Y)
	b := 200 * (var_Y - var_Z)

	return l,a,b
}

func RGBToLAB(color prominentcolor.ColorItem) (float64, float64, float64) {
	x,y,z := RGBtoXYZ(color)
	l,a,b := XYZToLAB(x,y,z)
	return l,a,b
}

func RGBToHSL(color prominentcolor.ColorItem) (float64, float64, float64) {
	var_R := float64(color.Color.R) / 255
	var_G := float64(color.Color.G) / 255
	var_B := float64(color.Color.B) / 255

	var_Min := math.Min(var_R, var_G)
	var_Min = math.Min(var_Min, var_B)
	var_Max := math.Max(var_R, var_G)
	var_Max = math.Max(var_Max, var_B)
	del_Max := var_Max - var_Min

	l := float64((var_Max + var_Min)) / 2
	h := float64(0)
	s := float64(0)

	if (del_Max != 0) {
		if (l < 0.5) {
			s = del_Max / (var_Max + var_Min)
		} else {
			s = del_Max / (2 - var_Max - var_Min)
		}

		del_R := (((float64(var_Max - var_R)) / 6) + (float64(del_Max) / 2)) / del_Max
		del_G := (((float64(var_Max - var_G)) / 6) + (float64(del_Max) / 2)) / del_Max
		del_B := ((float64((var_Max - var_B) / 6)) + (float64(del_Max) / 2)) / del_Max

		if (var_R == var_Max) {
			h = del_B - del_G
		} else if (var_G == var_Max) {
			h = (float64(1) / 3) + del_R - del_B
		} else {
			h = (float64(2) / 3) + del_G - del_R
		}
	
		if (h < 0) {
			h += 1
		}

		if (h > 1) {
			h -= 1
		}
	}

	return h,s,l
}

func betterSimilarColor(color1 prominentcolor.ColorItem, color2 prominentcolor.ColorItem) float64 {
	color1_L, color1_a, color1_b := RGBToLAB(color1)
	color2_L, color2_a, color2_b := RGBToLAB(color2)

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

	// NEED TO CHECK
	if math.IsNaN(delta_h) {
		delta_h = 0
	}

	sl := 1
	sc := 1 + (k1 * c1)
	sh := 1 + (k2 * c1)

	delta_e := math.Sqrt(math.Pow((delta_L/float64(kl*sl)), 2) + math.Pow((delta_c/float64(kc*sc)), 2) + math.Pow((delta_h/(kh*sh)), 2))

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
