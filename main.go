package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	connectedUsers    = 0
	nameToConnections = make(map[string]*websocket.Conn)
	messageChannel    = make(chan message)
)

type message struct {
	msgType int
	name    string
	msg     string
}

func broadcastMessage() {
	for {
		msg := <-messageChannel
		for name, conn := range nameToConnections {
			if name != msg.name {
				conn.WriteMessage(msg.msgType, []byte(msg.name+" says: "+msg.msg))
			}
		}
	}
}

func chatHandler(writer http.ResponseWriter, request *http.Request) {
	name := request.Header.Get("name")
	connectedUsers++
	if name == "" {
		name = fmt.Sprintf("User %d", connectedUsers)
	}
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		panic(err)
	}
	nameToConnections[name] = conn
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(msg))
		messageChannel <- message{
			msgType: msgType,
			name:    name,
			msg:     string(msg),
		}
	}
}

func main() {
	go broadcastMessage()
	http.HandleFunc("/chat", chatHandler)
	http.ListenAndServe(":8080", nil)
}
