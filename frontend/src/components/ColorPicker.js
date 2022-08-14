import React, {useState} from 'react'

import { HexColorPicker } from "react-colorful";

const ColorPicker = (text) => {
    const [color, setColor] = useState("#aabbcc");

    console.log(color)
    console.log(text)

    return (
        <>
            <HexColorPicker color={color} onChange={setColor} />
            <h2>{text.text} Color</h2>
        </>
    )
}

export default ColorPicker