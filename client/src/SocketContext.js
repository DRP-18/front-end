import React, { createContext, useState, useRef, useEffect } from "react";
import { io } from "socket.io-client";
import Peer from "simple-peer";

const SocketContext = createContext();

/* Pass full URL of deployed server*/
const socket = io("https://tutor-me-drp.herokuapp.com/");

const ContextProvider = ({ children }) => {
    /*These are state fields */
    const [stream, setStream] = useState(null);
    const [me, setMe] = useState("");
    const [call, setCall] = useState({});
    const [callAccepted, setCallAccepted] = useState(false);
    const [callEnded, setCallEnded] = useState(false);
    const [name, setName] = useState("");


    const myVideo = useRef();
    const userVideo = useRef();
    const connectionRef = useRef();


    useEffect(() => {
        /* Get permission for microphone and webcam*/
        navigator.mediaDevices.getUserMedia({ video: true, audio: true }) /* returns a promise*/
            .then((currentStream) => {
                setStream(currentStream);
                myVideo.current.srcObject = currentStream;
            });
        /* Listen for me action and get socketID*/
        socket.on("me", (id) => setMe(id));

        socket.on("calluser", ({ from, name: callerName, signal }) => {
            setCall({ isReceivedCall: true, from, name: callerName, signal })
        })
    }, []); /* Has an empty dependancy array*/


    const answerCall = () => {
        setCallAccepted(true);
        /*simple peer library usage */
        /* Initiator is who starts call
            stream from earlier getUserMedia
        */
        const peer = new Peer({ initiator: false, trickle: false, stream });

        peer.on("signal", (data) => {
            socket.emit("answercall", { signal: data, to: call.from })
            console.log("answercall sent")
        });
        peer.on("stream", (currentStream) => {
            /* This is the other persons stream*/
            userVideo.current.srcObject = currentStream;
        });

        peer.signal(call.signal);

        connectionRef.current = peer;

    }

    const callUser = (id) => {
        /*we are the person calling */
        const peer = new Peer({ initiator: true, trickle: false, stream });

        peer.on("signal", (data) => {
            socket.emit("calluser", { userToCall: id, signalData: data, from: me, name })
            console.log("calluser sent")
            console.log(me)
        });
        peer.on("stream", (currentStream) => {
            userVideo.current.srcObject = currentStream;
        });


        socket.on("callaccepted", (signal) => {
            setCallAccepted(true);
            peer.signal(signal);
        });

        connectionRef.current = peer;
    }

    const leaveCall = () => {
        setCallEnded(true);
        connectionRef.current.destroy(); /*Stop recieving input from user camera and microphone */

        window.location.reload();
    }

    return (
        /*This exposes all the information in this file to the package */
        <SocketContext.Provider value = {{call,callAccepted, myVideo, userVideo, stream, name, setName, callEnded, me, callUser, leaveCall, answerCall}}>
            {children}
        </SocketContext.Provider >
    );

}

export {ContextProvider, SocketContext};

