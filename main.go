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
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
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
	msgStr := strings.TrimSpace(string(msg))
	log.Printf("[%s]: %s", c.UserName, msgStr)

	if strings.HasPrefix(msgStr, "/create ") {
		newRoom := strings.TrimSpace(strings.TrimPrefix(msgStr, "/create "))
		if newRoom == "" {
			c.Send <- []byte("Использование: /create <имя_комнаты>")
		} else if b.CreateRoom(newRoom) {
			c.Send <- []byte("Комната создана: " + newRoom)
		} else {
			c.Send <- []byte("Комната уже существует: " + newRoom)
		}
		return
	}

	switch {
	case msgStr == "/who":
		users := b.GetUsernames(room)
		reply := "Сейчас в комнате: " + strings.Join(users, ", ")
		c.Send <- []byte(reply)

	case msgStr == "/rooms":
		rooms := b.GetRooms()
		reply := "Доступные комнаты: " + strings.Join(rooms, ", ")
		c.Send <- []byte(reply)

	case strings.HasPrefix(msgStr, "/join "):
		newRoom := strings.TrimSpace(strings.TrimPrefix(msgStr, "/join "))
		if newRoom != room {
			b.Leave(room, c)
			b.Broadcast(room, []byte(c.UserName + " вышел из комнаты"))

			room = newRoom
			b.Join(room, c)
			b.Broadcast(room, []byte(c.UserName + " вошел в комнату"))

			c.Send <- []byte("Вы вошли в комнату: " + room)
		}

	default:
		fullMsg := []byte(c.UserName + ": " + msgStr)
		b.Broadcast(room, fullMsg)
	}
	})
}

func main() {
	b := broker.New()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleConnection(w, r, b)
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("chat-server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
