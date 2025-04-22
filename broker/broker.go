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
	log.Printf("%s присоединился к комнате %s", c.UserName, room)
}

func (b *Broker) Leave(room string, c *client.Client) {
	if roomInterf, ok := b.rooms.Load(room); ok {
		roomMap := roomInterf.(*sync.Map)
		roomMap.Delete(c.UserName)
	}
	log.Printf("%s вышел из комнаты %s", c.UserName, room)
	b.cleanupRoom(room)
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

func (b *Broker) GetRooms() []string {
	var rooms []string
	b.rooms.Range(func(key, _ any) bool {
		roomName, ok := key.(string)
		if ok {
			rooms = append(rooms, roomName)
		}
		return true
	})
	return rooms
}


func (b *Broker) CreateRoom(room string) bool {
	_, loaded := b.rooms.LoadOrStore(room, &sync.Map{})
	if !loaded {
		log.Printf("Комната %q была создана", room)
		return true
	}
	return false
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


func (b *Broker) cleanupRoom(room string) {
	if room_interf, ok := b.rooms.Load(room); ok {
		roomMap := room_interf.(*sync.Map)
		var count int 
		roomMap.Range(func(key, value any) bool {
			count++
			return true
		})

		if count == 0 {
			b.rooms.Delete(room)
			log.Printf("Комната %s была закрыта", room)
		}
	}
}