import React, {useEffect, useState} from "react";
import './assets/text.css';
import Layout from "./components/Layout";
import AlbumSearch from "./components/AlbumSearch";
import { Link } from "react-router-dom";

const App = () => {
  const [backgroundColor, setBackgroundColor] = useState("#00d3ff");

  const getBackgroundImage = (data) => {
    console.log(data);
    setBackgroundColor(data);
  }

  return (
    <div className="h-full min-h-screen" style={{backgroundImage: `linear-gradient(to bottom, ${backgroundColor}, white)`}}>
      <Layout>
          <AlbumSearch func={getBackgroundImage}/>
      </Layout>
    </div>
  );
}

export default App;
