// Define "require"
import { createRequire } from "module";
const require = createRequire(import.meta.url);

var express = require('express'); // Express web server framework
var request = require('request-promise'); // "Request" library
const url = require('url');
var path = require('path');
var cors = require('cors');
var querystring = require('querystring');
const { createCanvas, loadImage } = require('canvas')
var bodyParser = require('body-parser')
var cookieParser = require('cookie-parser');
var getColors = require('get-image-colors');
import hexRgb from "hex-rgb";
require('dotenv').config()
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
import { fileURLToPath } from 'url';
// const { createProxyMiddleware } = require('http-proxy-middleware');
import { log } from "console";

console.log(process.env.CLIENT_ID + " + " + process.env.CLIENT_SECRET);

var client_id = process.env.CLIENT_ID; // Your client id
var client_secret = process.env.CLIENT_SECRET; // Your secret
var redirect_uri = 'http://localhost:8888/callback'; // Your redirect uri
var port = 8888;
var token = "";
var artist_api

var app = express();

// parse application/x-www-form-urlencoded
app.use(bodyParser.urlencoded({ extended: false }))
// parse application/json
app.use(bodyParser.json())
app.use(cors())
  .use(cookieParser())
  .use(express.static(path.join(__dirname, "../client/build")));

// export default function (app) {
//   app.use(
//     '/api',
//     createProxyMiddleware({
//       target: 'http://localhost:8888',
//       changeOrigin: true,
//     })
//   );
// };

// Handle GET requests to /api route
app.get("/api", (req, res) => {
  res.sendStatus(200);
});

/**
 * Endpoint which retrieves data for the album that was been pasted into the search bar
 */
app.post('/api/get-album', async (req, res) => {

  var uri = req.body.album_url.split(':')[2]

  var options = {
    url: `https://api.spotify.com/v1/albums/${uri}`,
    headers: { 'Authorization': 'Bearer ' + token },
    json: true
  };

  request.get(options, async function (error, response, body) {
    var artist = body.artists[0].name
    var image = body.images[0].url
    var album = body.name

    artist_api = body.artists[0].href

    const options = {
      count: 10,
    }

    var colors = await getColors(image, options)
    colors = colors.map(color => color.hex())

    var proportionColors = await colorProportions(colors, body.images[0].url)

    var result = []

    proportionColors.forEach((value, key) => {
      result.push([key, value])
    })
    
    var proportions = []

    for (var i = 0; i < result.length; i++) {
      colors[i] = result[i][0]
      proportions[i] = result[i][1]
    }

    res.redirect(url.format({
      pathname: "/related-artists",
      query: {
        "artist": artist,
        "album_image": image,
        "album": album,
        "imageColors": colors,
        "artist_info": artist_api,
        "proportions": proportions
      }
    }));
  });
});

/**
 * Endpoint which retrieves data for the album that was been pasted into the search bar
 */
app.post('/api/get-album', (req, res) => {

  var uri = req.body.album_url.split('/')[4].substring(0, 22)
  console.log(uri)
  var options = {
    url: `https://api.spotify.com/v1/albums/${uri}`,
    headers: { 'Authorization': 'Bearer ' + token },
    json: true
  };

  request.get(options, function (error, response, body) {
    if (response.statusCode != 404) {
      console.log(options)
      getAlbumInfo(body, res, url)
    }
    else {
      options.url = `https://api.spotify.com/v1/tracks/${uri}`;
      request.get(options, function (error, response, body) {
        options.url = `https://api.spotify.com/v1/albums/${body.album.uri.split(':')[2]}`
        console.log(options)
        request.get(options, function (error, response, body) {
          console.log(body)
          getAlbumInfo(body, res, url)
        })
      }) 
    }
  });
})

app.get("/related-artists", (req, res) => {

  var options = {
    url: req.query.artist_info + "/related-artists",
    headers: { 'Authorization': 'Bearer ' + token },
    json: true
  };

  request.get(options, (error, response, body) => {
    var relatedArtists = {
      artist: req.query.artist,
      album_image: req.query.album_image,
      album: req.query.album,
      imageColors: req.query.imageColors,
      occurencesOfColor: req.query.proportions,
      artists: [],
      uris: []
    }

    for (var i = 0; i < body.artists.length; i++) {
      relatedArtists.artists[i] = body.artists[i].name
      relatedArtists.uris[i] = body.artists[i].uri
    }

    console.log(relatedArtists)
    res.json(relatedArtists)
  })

})

app.post("/api/retrieve-new-music", async (req, res) => {
  var uris = req.body.relatedURIs
  var originalColorScheme = req.body.colorScheme
  originalColorScheme = originalColorScheme.filter(color => hexRgb(color))
  var originalColorProportions = req.body.colorProportions

  var albumsToReturn = await retrieveAlbums(uris, token, originalColorScheme, originalColorProportions)

  res.json(albumsToReturn)
})

async function retrieveAlbums(uris, token, originalColorScheme, originalColorProportions) {
  var albumsToReturn = []
  var visitedAlbums = new Set()

  // For each artist
  for (let i = 0; i < uris.length; i++) {
    var id = uris[i].split(':')[2]

    var options = {
      url: `https://api.spotify.com/v1/artists/${id}/albums`,
      headers: { 'Authorization': 'Bearer ' + token },
      json: true
    };

    // Get albums from that artist

    let data = await request(options)

    // For each album
    for (const item of data.items) {
      if (item.album_type == "album") {
        var artists = []
        for (var j = 0; j < item.artists.length; j++) {
          artists[j] = item.artists[j].name
        }


        const color_options = {
          count: 6,
        }

        //Get color scheme of the album
        let colors = await getColors(item.images[0].url, color_options)

        if (!colors) {
          throw colors
        }

        let count = 0;

        colors = colors.map(color => color.hex())

        var proportionColors = await colorProportions(colors, item.images[0].url)

        //Compare each color in original album with the color you're looking at
        proportionColors.forEach((value, key) => {
          var colorDiff = rgbDifference(key, originalColorScheme, 25)
          if (colorDiff[1]) { //if true
            var indexOfColor = originalColorScheme.indexOf(colorDiff[0])
            if (similarProportions(value, originalColorProportions[indexOfColor])) {
              count++;
            }
          }
        })

        // Add to list if 90% of colours match
        if ((count / colors.length) >= 0.8 && !(visitedAlbums.has(item.name))) {
          console.log("colour match: " + (count / colors.length) * 100 + "%");

          var relatedAlbum = {
            type: item.album_type,
            id: item.id,
            name: item.name,
            artists: artists,
            image: item.images[0].url

          }
          console.log(relatedAlbum);
          albumsToReturn.push(relatedAlbum)
          visitedAlbums.add(relatedAlbum.name)
        }
      }
    }

    console.log(i);
  }

  console.log("end of loop");

  return albumsToReturn;
}

function rgbDifference(color, originalColorScheme, difference) {
  //Loop through originalColourScheme
  for (const originalColor of originalColorScheme) {

    //Convert original colour and new colour to RGB
    const originalColorRGB = hexRgb(originalColor)
    const currentColorRGB = hexRgb(color)

    // Create heuristic ranges for red,green and blue values based on original colour
    const range = {
      red: {
        max: originalColorRGB.red + difference,
        min: originalColorRGB.red - difference
      },
      green: {
        max: originalColorRGB.green + difference,
        min: originalColorRGB.green - difference
      },
      blue: {
        max: originalColorRGB.blue + difference,
        min: originalColorRGB.blue - difference
      }
    }

    //calculate the difference in rgb.
    if (currentColorRGB.red <= range.red.max && currentColorRGB.red >= range.red.min) {
      if (currentColorRGB.green <= range.green.max && currentColorRGB.green >= range.green.min) {
        if (currentColorRGB.blue <= range.blue.max && currentColorRGB.blue >= range.blue.min) {
          //The difference is withn the specified metric
          return [originalColor, true]
        }
      }
    }
  }

  return ['', false]
}

function similarProportions(occurences, originalColorOccurences) {

    // 30% difference
    var percentageDifference = originalColorOccurences * 0.3 
    var max = originalColorOccurences + percentageDifference
    var min = originalColorOccurences - percentageDifference

    if (occurences <= max && occurences >= min){
      return true
    }else{
      return false
    }

}

async function colorProportions(colors, imageUrl) {

  //construct a dict of key->color value->percentage of image
  let proportions = new Map();

  for (let i = 0; i < colors.length; i++) {
    proportions.set(colors[i], 0)
  }

  //We use a canvas to draw our image and look at each pixel through metadata

  const canvas = createCanvas(640, 640)
  const ctx = canvas.getContext('2d')

  var image = await loadImage(imageUrl)
  ctx.drawImage(image, 0, 0, 640, 640)

  const scannedImage = ctx.getImageData(0, 0, canvas.width, canvas.height)

  // go through each pixel in the image 
  for (let i = 0; i < scannedImage.data.length; i += 4) {
    //get rgb values 
    const r = scannedImage.data[i]
    const g = scannedImage.data[i + 1]
    const b = scannedImage.data[i + 2]

    //for each color in proportions map
    proportions.forEach(function (value, key) {
      var colorRgb = hexRgb(key)

      const range = {
        red: {
          max: colorRgb.red + 10,
          min: colorRgb.red - 10
        },
        green: {
          max: colorRgb.green + 10,
          min: colorRgb.green - 10
        },
        blue: {
          max: colorRgb.blue + 10,
          min: colorRgb.blue - 10
        }
      }

      // if a colour is close enough then increment the occurences counter for that color
      if (r <= range.red.max && r >= range.red.min) {
        if (g <= range.green.max && g >= range.green.min) {
          if (b <= range.blue.max && b >= range.blue.min) {
            proportions.set(key, proportions.get(key) + 1)
          }
        }
      }

    })

  }

  //sort map by 'dominant' colours
  var sortedProportions = new Map([...proportions.entries()].sort((a, b) => b[1] - a[1]))

  return sortedProportions
}

// All other GET requests not handled before will return our React app
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, '../client/build', 'index.html'));
});

var authOptions = {
  url: 'https://accounts.spotify.com/api/token',
  headers: {
    'Authorization': 'Basic ' + (new Buffer.from(client_id + ':' + client_secret).toString('base64'))
  },
  form: {
    grant_type: 'client_credentials'
  },
  json: true
};

request.post(authOptions, function (error, response, body) {
  if (!error && response.statusCode === 200) {
    token = body.access_token;
  }
});

app.listen(8888, () => {
  console.log(`Hueify listening on port 8888`)
})
