import Colors from './Colors'
import React, {useEffect, useState} from "react";
import useScript from '../assets/hooks/useScript';
import Text from "./Text";

const Album = (props) => {
  const {
      uri,
    image,
    name,
    artist,
    colors,
} = props

    const [openPlayer, setOpenPlayer] = useState(false);

    // useEffect(() => {
    //     document.
    // },[])
    //                  // onMouseLeave={e => setOpenPlayer(false)}

  return (
    <div className="flex flex-col items-center justify-center mt-6">
      <div className="h-64 w-64 border-1.5 drop-shadow-lg border-gray-900">
          { openPlayer ?
              <iframe
                  onMouseLeave={e => setOpenPlayer(false)}
                  style={{borderRight: "12px"}}
                  src={`https://open.spotify.com/embed/album/${ uri }?utm_source=generator`}
                  width="100%"
                  frameBorder="0" allowFullScreen=""
                  height="352"
                  allow="clipboard-write; encrypted-media; picture-in-picture"
                  loading="lazy">
              </iframe>
              :
              <img onMouseOver={e => setOpenPlayer(true)} src={image} className="rounded-md" alt="album-cover" />
              <h1 className="font-blinker mt-3 font-semibold text-lg">{name}</h1>
              <h3 className="font-blinker font-semibold">{artist}</h3>
              <Colors colors={colors} />
          }
      </div>
    </div>
  )
}

export default Album;
