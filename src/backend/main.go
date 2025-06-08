package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

// Message wraps a payload with its origin
type Message struct {
	data   []byte
	sender *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
	lock      = sync.Mutex{}
)

// Get message and broadcast it to everyone
func handleMessages() {
	for {
		msg := <-broadcast
		lock.Lock()
		for client := range clients {
			if client == msg.sender {
				continue
			}
			err := client.WriteMessage(websocket.TextMessage, msg.data)
			if err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
		lock.Unlock()
	}
}

// handleConnection reads messages and forwards them into the broadcast channel
func handleConnection(conn *websocket.Conn) {
	defer func() {
		lock.Lock()
		delete(clients, conn)
		lock.Unlock()
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		fmt.Println("Received:", string(msg))
		broadcast <- Message{data: msg, sender: conn}
	}
}

// serveWs upgrades HTTP requests to WebSocket connections
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	lock.Lock()
	clients[ws] = true
	lock.Unlock()

	go handleConnection(ws)
}

// setupRoutes registers HTTP handlers
func setupRoutes() {
	http.HandleFunc("/ws", serveWs)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("src/frontend"))))
}

func main() {
	setupRoutes()	// register /ws and file-server handler
	go handleMessages()
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

