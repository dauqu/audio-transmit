package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{} // WebSocket upgrader

type Room struct {
	ID           string
	Clients      map[*websocket.Conn]bool
	ClientsMutex sync.Mutex
}

var rooms map[string]*Room
var roomsMutex sync.Mutex

func createRoom(roomID string) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	room := &Room{
		ID:           roomID,
		Clients:      make(map[*websocket.Conn]bool),
		ClientsMutex: sync.Mutex{},
	}
	rooms[roomID] = room
}

func deleteRoom(roomID string) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	delete(rooms, roomID)
}

func joinRoom(roomID string, conn *websocket.Conn) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	room, ok := rooms[roomID]
	if !ok {
		log.Printf("Room '%s' not found\n", roomID)
		return
	}

	room.ClientsMutex.Lock()
	defer room.ClientsMutex.Unlock()

	room.Clients[conn] = true
	log.Printf("Client joined room '%s'\n", roomID)
}

func leaveRoom(roomID string, conn *websocket.Conn) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	room, ok := rooms[roomID]
	if !ok {
		log.Printf("Room '%s' not found\n", roomID)
		return
	}

	room.ClientsMutex.Lock()
	defer room.ClientsMutex.Unlock()

	delete(room.Clients, conn)
	log.Printf("Client left room '%s'\n", roomID)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Read room ID and join the room
	_, roomID, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	joinRoom(string(roomID), conn)

	// Read and write messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Broadcast message to all clients in the room
		room, ok := rooms[string(roomID)]
		if !ok {
			log.Printf("Room '%s' not found\n", roomID)
			break
		}

		room.ClientsMutex.Lock()
		for client := range room.Clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
				continue
			}
		}
		room.ClientsMutex.Unlock()
	}

	leaveRoom(string(roomID), conn)
}

func main() {
	rooms = make(map[string]*Room)
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
