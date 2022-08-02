import React from 'react'

const Colors = (props) => {
  const {
    colors
  } = props

  console.log(colors[0].Color.R)

  return (
    <div className="flex justify-center mt-2">
      {typeof(colors) !== "string" && (colors.map(color => {
        return <div className="h-5 w-5" style={{"backgroundColor": `rgba(${color.Color.R}, ${color.Color.G}, ${color.Color.B}, 1)`}} />
      }))}
    </div>
  )
}

export default Colors;
