package scaling

import (
	"pandagame/internal/config"
	"pandagame/internal/framework"
	"pandagame/internal/redisconn"
	"slices"
)

type redisRooms struct {
	conn redisconn.RedisConn
}

func NewRedisRooms(cfg config.RedisConfig) framework.Rooms {
	return &redisRooms{conn: redisconn.NewRedisConn(cfg)}
}

func (r *redisRooms) AddToRoom(connId string, roomId string) {
	members := r.RoomMembers(roomId)
	if !slices.Contains(members, connId) {
		members = append(members, connId)
		redisconn.SetThing(roomId, &members, r.conn)
	}
}

func (r *redisRooms) RemoveFromRoom(connId string, roomId string) {
	members := r.RoomMembers(roomId)
	i := slices.Index(members, connId)
	if i >= 0 {
		members = slices.Delete(members, i, i+1)
		redisconn.SetThing(roomId, &members, r.conn)
	}
}

func (r *redisRooms) RoomMembers(roomId string) []string {
	members, err := redisconn.GetThing[[]string](roomId, r.conn)
	if err != nil {
		return nil
	}
	return *members
}
