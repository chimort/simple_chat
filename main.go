package main

import (
	"chatserv/broker"
	"chatserv/client"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleConnection(w http.ResponseWriter, r *http.Request, b *broker.Broker) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ошибка апгрейда:", err)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "null"
	}

	c := client.New(conn, username)
	room := "default"
	log.Printf("%s вошел в комнату %s\n", username, room)

	b.Join(room, c)
	b.Broadcast(room, []byte(c.UserName+" Вошел в комнату!"))
	defer func() {
		b.Leave(room, c)
		b.Broadcast(room, []byte(c.UserName+" Вышел из комнаты"))
		log.Printf("%s покинул комнату %s", c.UserName, room)
	}()

	go c.WritePump()

	c.ReadPump(func(msg []byte) {
		msgStr := string(msg)
		log.Printf("[%s]: %s", c.UserName, msgStr)

		if msgStr == "/who\n" {
			users := b.GetUsernames(room)
			reply := "Сейчас в чате: " + strings.Join(users, ", ")
			c.Send <- []byte(reply)
			return
		}

		fullMsg := []byte(c.UserName + ": " + msgStr)
		b.Broadcast(room, fullMsg)
	})
}

func main() {
	b := broker.New()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleConnection(w, r, b)
	})

	log.Println("chat-server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
