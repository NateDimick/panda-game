package framework

import (
	"slices"
	"sync"
)

type Rooms interface {
	AddToRoom(connId string, roomId string)
	RemoveFromRoom(connId string, roomId string)
	RoomMembers(roomId string) []string
}

type inMemRooms struct {
	rooms map[string][]string
	lock  sync.RWMutex
}

func NewInMemStorage() Rooms {
	ims := new(inMemRooms)
	ims.rooms = make(map[string][]string)
	return ims
}

func (i *inMemRooms) AddToRoom(id, roomId string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	members, ok := i.rooms[roomId]
	if ok {
		if slices.Contains(members, id) {
			return
		}
		i.rooms[roomId] = append(members, id)
	} else {
		i.rooms[roomId] = []string{id}
	}
}

func (i *inMemRooms) RemoveFromRoom(id, roomId string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	members, ok := i.rooms[roomId]
	if !ok {
		return
	}
	if !slices.Contains(members, id) {
		return
	}
	position := slices.Index(members, id)
	i.rooms[roomId] = slices.Delete(members, position, position+1)
}

func (i *inMemRooms) RoomMembers(roomId string) []string {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.rooms[roomId]
}
