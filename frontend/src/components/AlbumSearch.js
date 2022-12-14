import React, {useEffect, useState} from "react";
import Album from "./Album";
import axios from "axios";
import {Link} from "react-router-dom";
// import SSE from 'sse';
import Text from "./Text";

const AlbumSearch = () => {

    const [url, setUrl] = useState("");
    const [album, setAlbum] = useState({})
    const [recommendedAlbums, setRecommendedAlbums] = useState([])
    // const [recommendedAlbum, setRecommendedAlbum] = useState({})
    const [goClicked, setGoClicked] = useState(false)
    const [renderNewAlbums, setNewRender] = useState(false)

    useEffect(() => {
        if (Object.keys(album).length !== 0){
            console.log(album)
            getRecommendedAlbums()
        }
    }, [album])

    const getAlbum = (event) => {
        event.preventDefault();

        setGoClicked(true)
        axios.get(`/api/get-album/${url}`)
            .then((res) => {
                setAlbum(res.data)
            })
    }

    const getRecommendedAlbums = () => {

        const source = new EventSource(`/api/recommended/${album.album_id}/${album.artist_id}`)

        source.onmessage = (event) => {
            const jsonData = JSON.parse(event.data);
            setRecommendedAlbums(prevArray => [...prevArray, jsonData]);
            if (jsonData.endStream === true){
                source.close()
            }
        }
    }

    useEffect(() => {
        console.log(recommendedAlbums)
        setNewRender(true);
    }, [recommendedAlbums]);

    return (
        <>
        <div className="flex justify-center mt-8">
            <div className="mb-3 xl:w-96">
                <div className="input-group relative flex items-stretch w-full mb-4">
                    <input type="search" className="form-control relative flex-auto min-w-0 block w-full px-3 py-1.5 text-base font-normal text-gray-700 bg-white bg-clip-padding border border-solid border-gray-300 rounded transition ease-in-out m-0 focus:text-gray-700 focus:bg-white focus:border-blue-600 focus:outline-none" placeholder="Input song or album link" aria-label="Search" aria-describedby="button-addon3" value={url} onChange={e => setUrl(e.target.value)} />
                    <button className="btn inline-block px-6 py-2 border-2 border-gray-900 bg-gray-900 text-gray-50 font-medium text-xs leading-tight uppercase rounded hover:bg-black focus:outline-none focus:ring-0 transition duration-150 ease-in-out" type="button" id="button-addon3" onClick={getAlbum}>GO</button>
                </div>
            </div>
        </div>
        {!goClicked
            ?
            <div className="flex justify-center ">
                <button className="btn inline-block px-6 py-2 border-2 border-gray-900 bg-gray-900 text-gray-50 font-medium text-xs leading-tight uppercase rounded hover:bg-black focus:outline-none focus:ring-0 transition duration-150 ease-in-out">
                    <Link to={`/create`}>Create Playlist</Link>
                </button>
            </div>
            :
            <></>
        }
        {Object.keys(album).length !== 0 &&
        <div>
            <Album image={album.album_image} name={album.album_name} artist={album.artist} colors={album.image_colors} />
            {album.new_request === true ? <Text text={"This is a new request"} loading={album.new_request}/> : <></>}
        </div>
        }
        {renderNewAlbums
            ?
            <div style={{display: "flex"}}>
                {
                    recommendedAlbums.map((rec) =>
                        <div>
                            <Album image={rec.image} name={rec.name} artist={rec.artists} colors={[]} />
                        </div>
                    )
                }
            </div>
            :
            <></>
        }
        </>
    )
}

export default AlbumSearch;