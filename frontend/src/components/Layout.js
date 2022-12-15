import React, {useState} from "react";
import logo from '../assets/hueify logo.png'
import '../assets/text.css';
import {Link} from "react-router-dom";

const Layout = ({children}) => {
  return (
    <div>
      <div className="flex justify-center pt-64">
          <Link to={`/`}>
            <img src={logo} className="w-64 h-20" alt="logo" />
          </Link>
      </div>
      <h2 className="flex justify-center mt-8 text-lg font-semibold font-blinker" style={{textShadow: "1px 1px 2px black"}}>
          Discover new music based on album artwork!
      </h2>
      <main>{children}</main>
    </div>
  );
}

export default Layout;
