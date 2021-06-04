package main

import (
	"fmt"
	"net/http"
	"os"
	"html/template"
	"log"
	"github.com/gorilla/websocket"
	socketio "github.com/googollee/go-socket.io"
	rice "github.com/GeertJohan/go.rice"
	"time"
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
		(writer).Header().Set("Access-Control-Allow-Origin", "*")
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

func setUpVideoChatSockets() *socketio.Server {
	server := socketio.NewServer(nil)

    server.OnConnect("/videoCall", func(so socketio.Conn) error {
		so.Emit("me", so.ID())
		return nil
	})

	server.OnEvent("/videoCall", "disconnect", func (so socketio.Conn) {
		server.BroadcastToNamespace("/videoCall", "callended")
	})

	server.OnEvent("/videoCall", "calluser", func (so socketio.Conn, userToCall, singalData, from, name string) {
		server.BroadcastToRoom(userToCall, singalData, from, name)
	})

	server.OnEvent("/videoCall", "answercall", func (so socketio.Conn, data struct {to string; signal string}) {
		server.BroadcastToRoom(data.to, "callaccepted", data.signal)
	})

	return server
}

func serveAppHandler(app *rice.Box) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexFile, err := app.Open("index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, "index.html", time.Time{}, indexFile)
	}
}

func main() {
	go broadcastMessage()

	server := setUpVideoChatSockets()
	go server.Serve()
	defer server.Close()
	http.Handle("/socket.io/", server)
	
	// Define the rice box with the frontend client static files.
	appBox, err := rice.FindBox("./client/build")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/static/", http.FileServer(appBox.HTTPBox()))
	http.Handle("/static2/", http.StripPrefix("/static2", http.FileServer(http.Dir("./website"))))
	http.HandleFunc("/", serveSimplePage("index.html"))
	http.HandleFunc("/textChat/chat", chatHandler)
	http.HandleFunc("/videoCall", serveAppHandler(appBox))
	// http.HandleFunc("/videoCall", serveSimplePage("videoCall.html"))
	http.HandleFunc("/voiceCall", serveSimplePage("voiceCall.html"))
	http.HandleFunc("/textChat", serveSimplePage("textChat.html"))
	http.ListenAndServe(GetPort(), nil)
}
