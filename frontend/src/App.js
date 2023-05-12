import React, {useEffect, useState} from "react";
import './assets/text.css';
import Layout from "./components/Layout";
import AlbumSearch from "./components/AlbumSearch";
import { Link } from "react-router-dom";

const App = () => {
  const [backgroundColor, setBackgroundColor] = useState(["#0053ff","#00d3ff"]);

  const getBackgroundImage = (data) => {
    console.log("data: " + data);
    setBackgroundColor(data);
  }

  // useEffect(()=>{
  //   setSecondaryColor();
  // },[primaryColor])

  return (
    <div className="h-full min-h-screen" style={{backgroundImage: `radial-gradient(${backgroundColor[0]}, ${backgroundColor[1]})`}}>
      <Layout>
          <AlbumSearch func={getBackgroundImage}/>
      </Layout>
    </div>
  );
}

export default App;
