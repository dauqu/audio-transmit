package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte)           // channel for broadcasting messages

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection to WebSocket:", err)
		return
	}

	clients[conn] = true // add client to connected clients map

	for {
		// Read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message from client:", err)
			delete(clients, conn)
			break
		}

		// Broadcast message to all clients
		broadcast <- msg
	}
}

func broadcastMessages() {
	for {
		// Get message from broadcast channel
		msg := <-broadcast

		// Send message to all connected clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Failed to write message to client:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	go broadcastMessages()
	log.Fatal(http.ListenAndServe(":9000", nil))
}
