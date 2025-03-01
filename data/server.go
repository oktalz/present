package data

import (
	"log"
	"sync"
	"time"
)

var muWS sync.RWMutex

type Server interface {
	Register(userID string, isAdmin bool, currentSlide int64) (ch chan Message, err error)
	Unregister(id string)
	Broadcast(msg Message)
	Send(id string, msg Message)
	Pool(msg Message)
}

func NewServer() *server { //revive:disable:unexported-return
	server := &server{
		clients:    make(map[string]chan Message),
		admins:     make(map[string]chan Message),
		pools:      make(map[string]map[string]string),
		updatePool: make(chan string),
	}
	go server.RateLimitBroadcastPool()

	return server
}

type server struct {
	clients    map[string]chan Message
	admins     map[string]chan Message
	pools      map[string]map[string]string
	updatePool chan string
}

//revive:disable:flag-parameter,function-length,cognitive-complexity,cyclomatic
func (s *server) Register(userID string, isAdmin bool, currentSlide int64) (ch chan Message, err error) {
	muWS.Lock()
	defer muWS.Unlock()
	ch = make(chan Message)
	if isAdmin {
		s.admins[userID] = ch
		log.Println("registered admin", userID)
	} else {
		s.clients[userID] = ch
		log.Println("registered", userID)
	}

	s.clients[userID] = ch
	go func() { //revive:disable:datarace
		ch <- Message{
			ID:     userID,
			Author: "SERVER",
			// Slides: data.Presentation(),
			Slide: int(currentSlide),
		}
		s.BroadcastPoolsToID(userID)
	}()
	return ch, nil
}

func (s *server) Unregister(id string) {
	muWS.Lock()
	defer muWS.Unlock()
	log.Println("unregistered", id)
	delete(s.clients, id)
	delete(s.admins, id)
}

func (s *server) Broadcast(msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	// log.Println("broadcast", msg.Author)
	for _, ch := range s.admins {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
	}
	for _, ch := range s.clients {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
	}
}

func (s *server) BroadcastSingle(msg Message, id string) {
	muWS.RLock()
	defer muWS.RUnlock()
	// log.Println("broadcast", msg.Author)
	ch, ok := s.admins[id]
	if ok {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
		return
	}

	ch, ok = s.clients[id]
	if ok {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
	}
}

func (s *server) BroadcastAdmins(msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	for _, ch := range s.admins {
		go func(ch chan Message, msg Message) {
			ch <- msg
		}(ch, msg)
	}
}

func (s *server) Send(id string, msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	ch, ok := s.clients[id]
	if ok {
		ch <- msg
	}
}

func (s *server) Pool(msg Message) {
	muWS.RLock()
	defer muWS.RUnlock()
	pool, ok := s.pools[msg.Pool]
	if ok {
		if pool[msg.Author] == msg.Value {
			return
		}
		pool[msg.Author] = msg.Value
	} else {
		s.pools[msg.Pool] = make(map[string]string)
		s.pools[msg.Pool][msg.Author] = msg.Value
	}
	// go s.BroadcastPool(msg.Pool)
	go func() {
		s.updatePool <- msg.Pool
	}()
}

func (s *server) RateLimitBroadcastPool() {
	poolsToUpdate := make(map[string]struct{})
	for {
		select {
		case pool := <-s.updatePool:
			poolsToUpdate[pool] = struct{}{}
		case <-time.After(150 * time.Millisecond):
			if len(poolsToUpdate) > 0 {
				for pool := range poolsToUpdate {
					go s.BroadcastPool(pool)
				}
				poolsToUpdate = make(map[string]struct{})
			}
		}
	}
}

func (s *server) BroadcastPool(pool string, ids ...string) {
	muWS.RLock()
	defer muWS.RUnlock()
	d := map[string]int{}
	for _, v := range s.pools[pool] {
		d[v]++
	}

	bMsg := Message{
		Pool: pool,
		Data: d,
	}
	if len(ids) == 0 {
		go s.Broadcast(bMsg)
	} else {
		for _, id := range ids {
			go s.BroadcastSingle(bMsg, id)
		}
	}
}

func (s *server) BroadcastPoolsToID(id string) {
	muWS.RLock()
	defer muWS.RUnlock()
	for k := range s.pools {
		go s.BroadcastPool(k, id)
	}
}
