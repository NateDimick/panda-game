package framework

import (
	"slices"
	"sync"
)

type Grouper interface {
	AddToGroup(connId string, groupId string)
	RemoveFromGroup(connId string, groupId string)
	GroupMembers(groupId string) []string
	DeleteGroup(groupId string)
}

type inMemGroups struct {
	groups map[string][]string
	lock   sync.RWMutex
}

func NewInMemStorage() Grouper {
	ims := new(inMemGroups)
	ims.groups = make(map[string][]string)
	return ims
}

func (i *inMemGroups) AddToGroup(id, groupId string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	members, ok := i.groups[groupId]
	if ok {
		if slices.Contains(members, id) {
			return
		}
		i.groups[groupId] = append(members, id)
	} else {
		i.groups[groupId] = []string{id}
	}
}

func (i *inMemGroups) RemoveFromGroup(id, groupId string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	members, ok := i.groups[groupId]
	if !ok {
		return
	}
	if !slices.Contains(members, id) {
		return
	}
	position := slices.Index(members, id)
	i.groups[groupId] = slices.Delete(members, position, position+1)
}

func (i *inMemGroups) GroupMembers(groupId string) []string {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.groups[groupId]
}

func (i *inMemGroups) DeleteGroup(groupId string) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	delete(i.groups, groupId)
}
