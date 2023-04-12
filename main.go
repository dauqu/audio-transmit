package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections (for demo purposes only)
		return true
	},
}

var roomsMutex sync.Mutex
var rooms = make(map[string]map[*websocket.Conn]bool)

func main() {

	// Static file server
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleWebSocket)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	defer conn.Close()

	for {
		// Read message from WebSocket
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage:", err)
			break
		}

		// Process message (e.g., join room or broadcast message)
		// Replace this with your own logic
		processMessage(conn, msg)
	}
}

func processMessage(conn *websocket.Conn, msg []byte) {
	// Parse the message as JSON
	var data map[string]interface{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		log.Println("Unmarshal:", err)
		return
	}

	// Check the type of message
	msgType, ok := data["type"].(string)
	if !ok {
		log.Println("Invalid message type")
		return
	}

	// Handle different types of messages
	switch msgType {
	case "join":
		// Extract the room ID from the message
		roomID, ok := data["room"].(string)
		if !ok {
			log.Println("Invalid room ID")
			return
		}

		// Add the connection to the room
		roomsMutex.Lock()
		if rooms[roomID] == nil {
			rooms[roomID] = make(map[*websocket.Conn]bool)
		}
		rooms[roomID][conn] = true
		roomsMutex.Unlock()

		fmt.Println("Client joined room:", roomID)

	case "message":
		// Extract the room ID and message from the message
		roomID, ok := data["room"].(string)
		if !ok {
			log.Println("Invalid room ID")
			return
		}

		message, ok := data["message"].(string)
		if !ok {
			log.Println("Invalid message")
			return
		}

		// Broadcast the message to all connections in the room
		roomsMutex.Lock()
		for conn := range rooms[roomID] {
			err := conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("WriteMessage:", err)
			}
		}
		roomsMutex.Unlock()

	default:
		log.Println("Unknown message type:", msgType)
	}
}
