import React, {useEffect, useState} from 'react'
import '../assets/text.css';

const Text = (props) => {
    const {
        text,
        loading
    } = props

    const [slides, setSlides] = useState(["text-box-slide-up","text-box-slide-down"])

    useEffect(() => {
        let timer = []
        for (let i = 0; i< slides.length; i++){
            setTimeout(() => {
            timer[i] = setSlides(previousValue => {
                let newValue = [...previousValue]
                newValue[i] = ''
                return newValue
            })}, 5000)
        }
        return () => {
            clearTimeout(timer)
        }
    }, [])

    return (
            slides.map((item, index) => (
                <div className={item} key={index}>
                    <span className={loading ? "font-blinker loading" : "font-blinker"}>{text}</span>
                </div>
            ))
    )
}

export default Text;