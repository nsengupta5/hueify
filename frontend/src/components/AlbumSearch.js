import React, { useEffect, useState } from "react";
import Album from "./Album";
import axios from "axios";
import { Link } from "react-router-dom";
// import SSE from 'sse';
import Text from "./Text";

const AlbumSearch = (props) => {
  const [url, setUrl] = useState("");
  const [album, setAlbum] = useState({})
  const [recommendedAlbums, setRecommendedAlbums] = useState([])
  // const [recommendedAlbum, setRecommendedAlbum] = useState({})
  const [goClicked, setGoClicked] = useState(false)
  const [renderNewAlbums, setNewRender] = useState(false)
  const [isSearching, setSearching] = useState(true)

  useEffect(() => {
    if (Object.keys(album).length !== 0) {
      getRecommendedAlbums()
    }
  }, [album])

  const getAlbum = (event) => {
    event.preventDefault();

    setGoClicked(true)

    if (url.includes("http")) {
      let transformedURL = ""

      if (url.includes("album")) {
        let albumSub = "album/"
        let startIdx = url.indexOf(albumSub) + albumSub.length
        let endIdx = url.indexOf("?")
        transformedURL = "spotify:album:" + url.substring(startIdx, endIdx)
      }
      else {
        let trackSub = "track/"
        let startIdx = url.indexOf(trackSub) + trackSub.length
        let endIdx = url.indexOf("?")
        transformedURL = "spotify:track:" + url.substring(startIdx, endIdx)
      }

      axios.get(`/api/get-album/${transformedURL}`)
        .then((res) => {
          setAlbum(res.data)
          let primaryColor = `rgba(${res.data.image_colors[0].Color.R},${res.data.image_colors[0].Color.G},${res.data.image_colors[0].Color.B},1)`
          let secondaryColor = `rgba(${res.data.image_colors[1].Color.R},${res.data.image_colors[1].Color.G},${res.data.image_colors[1].Color.B},1)`
          props.func([primaryColor, secondaryColor]);
        })
    }

    else {
      axios.get(`/api/get-album/${url}`)
        .then((res) => {
          setAlbum(res.data)
          let primaryColor = `rgba(${res.data.image_colors[0].Color.R},${res.data.image_colors[0].Color.G},${res.data.image_colors[0].Color.B},1)`
          let secondaryColor = `rgba(${res.data.image_colors[1].Color.R},${res.data.image_colors[1].Color.G},${res.data.image_colors[1].Color.B},1)`
          props.func([primaryColor, secondaryColor]);
        })
    }

  }

  const getRecommendedAlbums = () => {

    const source = new EventSource(`/api/recommended/${album.album_id}/${album.artist_id}`)

    source.onmessage = (event) => {
      const jsonData = JSON.parse(event.data);
      console.log(jsonData)
      setRecommendedAlbums(prevArray => [...prevArray, jsonData]);
      if (jsonData.endStream === true) {
        source.close()
        setSearching(false)
      }
    }
  }

  useEffect(() => {
    setNewRender(true);
  }, [recommendedAlbums]);

  return (
    <>
      <div className="flex justify-center mt-8">
        <div className="mb-3 xl:w-96">
          <div className="relative flex items-stretch w-full mb-4 input-group">
            <input type="search" className="form-control relative flex-auto min-w-0 block w-full px-3 py-1.5 text-base font-normal text-gray-700 bg-white bg-clip-padding border border-solid border-gray-300 rounded transition ease-in-out m-0 focus:text-gray-700 focus:bg-white focus:border-blue-600 focus:outline-none" placeholder="Input song or album link" aria-label="Search" aria-describedby="button-addon3" value={url} onChange={e => setUrl(e.target.value)} />
            <button className="inline-block px-6 py-2 text-xs font-medium leading-tight uppercase bg-gray-900 border-2 border-gray-900 rounded btn text-gray-50 hover:bg-black focus:outline-none focus:ring-0 transition duration-150 ease-in-out" type="button" id="button-addon3" onClick={getAlbum}>GO</button>
          </div>
        </div>
      </div>
      {!goClicked
        ?
        <div className="flex justify-center ">
          <button className="inline-block px-6 py-2 text-xs font-medium leading-tight uppercase bg-gray-900 border-2 border-gray-900 rounded btn text-gray-50 hover:bg-black focus:outline-none focus:ring-0 transition duration-150 ease-in-out">
            <Link to={`/create`}>Create Playlist</Link>
          </button>
        </div>
        :
        <></>
      }
      {Object.keys(album).length !== 0 &&
        <>
          <Album uri={album.album_id} image={album.album_image} name={album.album_name} artist={album.artist} colors={album.image_colors} />
          <div class="mt-4 mb-4">
            <div class="relative rounded-xl">
              <div class="flex justify-center items-center">
                <span class="inline-flex items-center px-4 py-2 leading-tight text-xs rounded text-black font-semibold">
                  {isSearching ?
                    <>
                      <svg class="animate-spin h-5 w-5 mr-3 text-black" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      <span>PROCESSING...</span>
                    </>
                    : <></>}
                </span>
              </div>
            </div>
          </div>
        </>
      }
      {renderNewAlbums &&
        <div class="grid grid-flow-row gap-4 lg:grid-cols-5 md:grid-cols-3 sm:grid-cols-1">
          {
            recommendedAlbums.map((rec) =>
              <div>
                <Album uri={rec.id} image={rec.image} name={rec.name} artist={rec.artists} colors={rec.colors} />
              </div>
            )
          }
        </div>
      }
    </>
  )
}

export default AlbumSearch;
