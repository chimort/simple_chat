package broker

import (
	"chatserv/client"
	"log"
	"sync"
)

type Broker struct {
	rooms sync.Map
}

func New() *Broker {
	return &Broker{}
}

func (b *Broker) Join(room string, c *client.Client) {
	roomInterf, _ := b.rooms.LoadOrStore(room, &sync.Map{})
	roomMap := roomInterf.(*sync.Map)

	roomMap.Store(c.UserName, c)
}

func (b *Broker) Leave(room string, c *client.Client) {
	if roomInterf, ok := b.rooms.Load(room); ok {
		roomMap := roomInterf.(*sync.Map)
		roomMap.Delete(c.UserName)
	}
}

func (b *Broker) GetUsernames(room string) []string {
	var users []string
	if roomInterf, ok := b.rooms.Load(room); ok {
		roomMap := roomInterf.(*sync.Map)
		roomMap.Range(func(key, value any) bool {
			if client, ok := value.(*client.Client); ok {
				users = append(users, client.UserName)
			} else {
				log.Println("не удалось привести значение к типу *client.Client")
			}
			return true
		})
	}
	return users
}

func (b *Broker) Broadcast(room string, msg []byte) {
	if roomInterf, ok := b.rooms.Load(room); ok {
		rm := roomInterf.(*sync.Map)
		rm.Range(func(key, value any) bool {
			cl, ok := value.(*client.Client)
			if !ok {
				log.Println("не удалось привести к типу *client.Client")
				return true
			}
			cl.Send <- msg
			return true
		})
	}
}
