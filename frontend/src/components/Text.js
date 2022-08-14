import React, {useEffect, useState} from "react";
import '../assets/text.css';

const Text = (props) => {
    const {
        text,
        loading
    } = props

    const [animation, setAnimation] = useState("text-box-slide-up")
    const [isLoading, setLoading] = useState(loading)
    const [newText, setText] = useState(text)

    useEffect(() => {
        const delay = t => new Promise(resolve => setTimeout(resolve, t));
        delay(5000).then(() => {
            setAnimation("text-box-slide-down");
            delay(1000).then(() =>{
                setAnimation("text-box-slide-up")
                setLoading(false)
                setText("Request Added 🎉")
                delay(3000).then(() =>{
                    setAnimation("fade-away")
                })
            })
        })
    }, [])

    return (
            <div className={animation}>
                <span className={isLoading ? "font-blinker loading" : "font-blinker"}>{newText}</span>
            </div>
    )
}

export default Text;