import React from "react";
import './assets/text.css';
import Layout from "./components/Layout";
import AlbumSearch from "./components/AlbumSearch";
import { Link } from "react-router-dom";

const App = () => {
  return (
    <Layout>
        <AlbumSearch/>
    </Layout>
  );
}

export default App;
