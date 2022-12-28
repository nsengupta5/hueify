import React, {useEffect, useState} from "react";
import Album from "./Album";
import axios from "axios";
import {Link} from "react-router-dom";
// import SSE from 'sse';
import Text from "./Text";

const AlbumSearch = (props) => {
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
                console.log(res.data.image_colors)
                let promColor = `rgba(${res.data.image_colors[0].Color.R},${res.data.image_colors[0].Color.G},${res.data.image_colors[0].Color.B},1)`
                props.func(promColor);
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
        <div>
            <Album uri={album.album_id} image={album.album_image} name={album.album_name} artist={album.artist} colors={album.image_colors} />
            {album.new_request === true ? <Text text={"This is a new request"} loading={album.new_request}/> : <></>}
        </div>
        }
        {renderNewAlbums
            ?
            <div className="grid grid-cols-4 grid-flow-row gap-4">
                {
                    recommendedAlbums.map((rec) =>
                        <div>
                            <Album uri={rec.id} image={rec.image} name={rec.name} artist={rec.artists} colors={rec.colors} />
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
