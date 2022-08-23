import React, { useState } from "react";
import logo from './assets/hueify logo.png';
import axios from 'axios'
import Album from './components/Album'

const App = () => {
  const [url, setUrl] = useState("");
  const [album, setAlbum] = useState("")

  const getAlbum = (event) => {
    event.preventDefault();
    axios.get(`/api/get-album/${url}`)
      .then((res) => {
        setAlbum(res.data)
      })
  }

  return (
    <div className="h-full min-h-screen bg-gradient-to-b from-blue-400 via-white-400 to-white">
      <div className="flex justify-center pt-64">
        <img src={logo} className="w-64 h-20" alt="logo" />
      </div>
      <h2 className="flex justify-center mt-8 text-lg font-medium font-blinker">Discover new music based on album artwork!</h2>
      <div className="flex justify-center mt-8">
        <div className="mb-3 xl:w-96">
          <div className="relative flex items-stretch w-full mb-4 input-group">
            <input type="search" className="form-control relative flex-auto min-w-0 block w-full px-3 py-1.5 text-base font-normal text-gray-700 bg-white bg-clip-padding border border-solid border-gray-300 rounded transition ease-in-out m-0 focus:text-gray-700 focus:bg-white focus:border-blue-600 focus:outline-none" placeholder="Input song or album link" aria-label="Search" aria-describedby="button-addon3" value={url} onChange={e => setUrl(e.target.value)} />
            <button className="inline-block px-6 py-2 text-xs font-medium leading-tight uppercase bg-gray-900 border-2 border-gray-900 rounded btn text-gray-50 hover:bg-black focus:outline-none focus:ring-0 transition duration-150 ease-in-out" type="button" id="button-addon3" onClick={getAlbum}>GO</button>
          </div>
        </div>
      </div>
      {album !== "" &&
      <div>
        <Album image={album.album_image} name={album.album_name} artist={album.artist} colors={album.image_colors} />
      </div>
      }
    </div>
  );
}

export default App;
