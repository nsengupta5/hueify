import React, {useState} from "react";
import logo from '../assets/hueifylogo3.png'
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
      <h2 className="flex justify-center mt-8 text-lg font-semibold font-blinker text-white">
          Discover new music based on album artwork!
      </h2>
      <main>{children}</main>
    </div>
  );
}

export default Layout;
