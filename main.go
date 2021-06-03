package main

import (
	"fmt"
	"net/http"
	"os"
	"html/template"
	"log"
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
				conn.WriteMessage(msg.msgType, []byte(msg.msg))
			}
		}
	}
}

func serveSimplePage(page string) func (writer http.ResponseWriter, request *http.Request) {
	return func (writer http.ResponseWriter, request *http.Request) {	
		template, err := template.ParseFiles(fmt.Sprintf("website/%s", page))
		if err != nil {
			log.Print("Error parsing template: ", err)
		}
		err = template.Execute(writer, nil)
		if err != nil {
			log.Print("Error during executing: ", err)
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

func GetPort() string {
	port := os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "8080"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

func main() {
	go broadcastMessage()
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./website"))))
	http.HandleFunc("/", serveSimplePage("index.html"))
	http.HandleFunc("/textChat/chat", chatHandler)
	http.HandleFunc("/videoCall", serveSimplePage("videoCall.html"))
	http.HandleFunc("/voiceCall", serveSimplePage("voiceCall.html"))
	http.HandleFunc("/textChat", serveSimplePage("textChat.html"))
	http.ListenAndServe(GetPort(), nil)
}
