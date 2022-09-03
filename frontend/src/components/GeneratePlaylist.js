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
        axios.post(`/api/create-playlist/`,
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
                <button className="btn inline-block px-6 py-2 border-2 border-gray-900 bg-gray-900 text-gray-50 font-medium text-xs leading-tight uppercase rounded hover:bg-black focus:outline-none focus:ring-0 transition duration-150 ease-in-out" type="button" id="button-addon3" onClick={genPlaylist}>Generate</button>
                <PlaylistCanvas />
            </div>
        </Layout>
    )
}

export default GeneratePlaylist
