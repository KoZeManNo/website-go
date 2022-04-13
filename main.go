package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/websocket"
)

var (
	//go:embed static
	files embed.FS

	prod         = flag.Bool("prod", false, "If the application is running on production")
	validMessage = regexp.MustCompile(`^[0-8][A-Fa-f0-9]{6}$`)
	upgrader     = websocket.Upgrader{}

	connections []*websocket.Conn
	colors      [9]string
)

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
	flag.Parse()

	for i := range colors {
		colors[i] = "000000"
	}

	fs, err := fs.Sub(files, "static")
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(fs)))
	http.HandleFunc("/ws", ws)
	if *prod {
		http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/kozemanno.dev/fullchain.pem", "/etc/letsencrypt/live/kozemanno.dev/privkey.pem", nil)
	} else {
		http.ListenAndServe(":8080", nil)
	}
}
