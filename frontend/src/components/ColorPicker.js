import React, {useEffect, useState} from "react";
import hexRgb from "hex-rgb"

import { HexColorPicker } from "react-colorful";

const ColorPicker = (props) => {
    const [color, setColor] = useState("#165da3");

    useEffect(() => {
        props.func(hexRgb(color), props.type)
    }, [color])

    console.log(color)
    console.log(props.text)

    return (
        <>
            <HexColorPicker color={color} onChange={setColor} />
            <h2>{props.text.text} Color</h2>
        </>
    )
}

export default ColorPicker
