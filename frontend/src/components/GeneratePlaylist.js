import React, {useState} from "react";
import ColorPicker from "./ColorPicker";
import Album from "./Album";
import axios from "axios";
import Layout from "./Layout"
import PlaylistCanvas from "./PlaylistCanvas";

const GeneratePlaylist = () =>{

    const [primaryColor, setPrimaryColor] = useState({})
    const [secondaryColor, setSecondaryColor] = useState({})
   
    const genPlaylist = (event) => {
        event.preventDefault();
        axios.post(`https://hueify.netlify.app/api/create-playlist/`,
            {
                primaryRGB: primaryColor,
                secondaryRGB: secondaryColor,
            })
            .then((res) => {
                console.log(res.data)
            })
    }

    const pullData = (data, isPrimary) => {
        if (isPrimary){
            setPrimaryColor(data)
        }else{
            setSecondaryColor(data)
        }
    }

    return (
        <Layout>
            <div className="flex justify-center pt-64">
                <ColorPicker text={"Primary"} func={pullData} type={true}/>
                <ColorPicker text={"Secondary"} func={pullData} type={false}/>
                <button className="inline-block px-6 py-2 text-xs font-medium leading-tight uppercase bg-gray-900 border-2 border-gray-900 rounded btn text-gray-50 hover:bg-black focus:outline-none focus:ring-0 transition duration-150 ease-in-out" type="button" id="button-addon3" onClick={genPlaylist}>Generate</button>
                <PlaylistCanvas />
            </div>
        </Layout>
    )
}

export default GeneratePlaylist
