import React from "react";
import logo from '../assets/hueify logo.png'
import '../assets/text.css';
import {Link} from "react-router-dom";

const Layout = ({children}) => {

  return (
    <div className="h-full min-h-screen bg-gradient-to-b from-cyan-400 via-white-400 to-white">
      <div className="flex justify-center pt-64">
          <Link to={`/`}>
            <img src={logo} className="h-20 w-64" alt="logo" />
          </Link>
      </div>
      <h2 className="flex font-blinker font-semibold text-lg justify-center mt-8" style={{textShadow: "1px 1px 2px black"}}>
          Discover new music based on album artwork!
      </h2>
      <main>{children}</main>
    </div>
  );
}

export default Layout;
