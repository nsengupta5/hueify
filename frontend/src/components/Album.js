import Colors from './Colors'
import React from "react";
import Text from "./Text";

const Album = (props) => {
  const {
    image,
    name,
    artist,
    colors
  } = props

  return (
    <div className="flex flex-col items-center justify-center mt-6">
      <div className="h-64 w-64 border-1.5 drop-shadow-lg border-gray-900">
        <img src={image} className="rounded-md" alt="album-cover" />
      </div>
      <h1 className="font-blinker mt-3 font-semibold text-lg">{name}</h1>
      <h3 className="font-blinker font-semibold">{artist}</h3>
      <Colors colors={colors} />
    </div>
  )
}

export default Album;