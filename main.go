package main

import (
	"embed"
	"io/fs"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/websocket"
)

//go:embed static
var files embed.FS
var validMessage = regexp.MustCompile(`^[0-8][A-Fa-f0-9]{6}$`)
var upgrader = websocket.Upgrader{}

var connections []*websocket.Conn
var colors [9]string

func removeConnection(conn *websocket.Conn) {
	for i, client := range connections {
		if client == conn {
			connections = append(connections[:i], connections[i+1:]...)
		}
	}
}

func ws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	connections = append(connections, conn)

	var all string
	for _, color := range colors {
		all += color
	}

	conn.WriteMessage(websocket.TextMessage, []byte(all))

	defer conn.Close()
	defer removeConnection(conn)

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if msgType == websocket.TextMessage {
			text := string(msg)
			if validMessage.MatchString(text) {
				index, err := strconv.Atoi(text[0:1])
				if err != nil {
					break
				}
				colors[index] = text[1:7]
				for _, client := range connections {
					client.WriteMessage(websocket.TextMessage, msg)
				}
			}
		}
	}
}

func main() {
	for i := range colors {
		colors[i] = "000000"
	}

	fs, err := fs.Sub(files, "static")
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(fs)))
	http.HandleFunc("/ws", ws)

	http.ListenAndServe(":8080", nil)
}
