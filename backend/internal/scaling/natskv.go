package scaling

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsKV struct {
	nc     *nats.Conn
	kv     jetstream.KeyValue
	bucket string
}

func NewNatsKV(addr, bucket string) *NatsKV {
	conn, err := nats.Connect(addr)
	if err != nil {
		slog.Warn("Failed to connect to nats", slog.String("error", err.Error()))
	}
	js, err := jetstream.New(conn)
	if err != nil {
		slog.Warn("failed to connect to jetstream", slog.String("error", err.Error()))
	}
	kv, err := js.KeyValue(context.Background(), bucket)
	if err != nil {
		slog.Warn("failed to setup jetstream kv", slog.String("error", err.Error()))
	}
	nk := &NatsKV{
		nc:     conn,
		bucket: bucket,
		kv:     kv,
	}

	return nk
}

func (n *NatsKV) AddToGroup(connId string, groupId string) {
	n.retryMembersModification(connId, groupId, func(members []string, connId string) []string {
		if slices.Contains(members, connId) {
			return nil
		}
		return append(members, connId)
	})
}

func (n *NatsKV) RemoveFromGroup(connId string, groupId string) {
	n.retryMembersModification(connId, groupId, func(members []string, connId string) []string {
		if !slices.Contains(members, connId) {
			return nil
		}
		i := slices.Index(members, connId)
		return slices.Delete(members, i, i+1)
	})
}

func (n *NatsKV) GroupMembers(groupId string) []string {
	members, _ := n.members(groupId)
	return members
}

func (n *NatsKV) DeleteGroup(groupId string) {
	n.kv.Delete(context.Background(), groupId)
}

func (n *NatsKV) members(groupId string) ([]string, uint64) {
	v, err := n.kv.Get(context.Background(), groupId)
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		rev, _ := n.kv.Create(context.Background(), groupId, []byte{'[', ']'})
		return make([]string, 0), rev
	}
	members := make([]string, 0)
	json.Unmarshal(v.Value(), &members)
	return members, v.Revision()
}

func (n *NatsKV) retryMembersModification(connId, groupId string, modifier func([]string, string) []string) {
	err := errors.New("placeholder")
	for err != nil {
		members, rev := n.members(groupId)
		members = modifier(members, connId)
		if members == nil {
			return
		}
		data, _ := json.Marshal(members)
		_, err = n.kv.Update(context.Background(), groupId, data, rev)
	}
}
