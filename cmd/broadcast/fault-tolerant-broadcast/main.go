// https://fly.io/dist-sys/3c/
package main

import (
	"context"
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
	"time"
)

// when we push a message, keep trying until its not longer in a a channel
// make 10 go routines to try and push messages to nodes instead of
// trying and then continuing
// or
// looping and blocking messages being received

type MultiNodeBroadcaster interface {
	Broadcast(msg maelstrom.Message) error
	Read(msg maelstrom.Message) error
	Topology(msg maelstrom.Message) error
}

type Server struct {
	n *maelstrom.Node
	m Messages
}

var mbnImplServer MultiNodeBroadcaster = &Server{}

func (s *Server) Broadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	value, ok := body["message"].(float64)
	if !ok {
		log.Printf("field 'message' is not an int64, got %T: %v", body["message"], body["message"])
	}

	// integer is always unique
	if !s.m.Exists(value) {
		s.m.Push(value)

		// tell adj nodes to update (this kinda ignores the topology)
		for _, dest := range s.n.NodeIDs() {
			if dest != msg.Dest && dest != msg.Src {
				// thing to perma try sending message (this is janky! wow!)
				go func() {
					ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
					// auto fail if not send within 1 second, retry until successful
					for {
						_, err := s.n.SyncRPC(ctx, dest, body)
						if err == nil {
							break
						}
					}
				}()
			}
		}
	}

	return s.n.Reply(msg, map[string]any{"type": "broadcast_ok"})
}

func (s *Server) Read(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	return s.n.Reply(msg, map[string]any{"type": "read_ok", "messages": s.m.Read()})
}

func (s *Server) Topology(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	return s.n.Reply(msg, map[string]any{"type": "topology_ok"})
}

type Messages struct {
	received map[int]struct{} // struct{} to minimize memory usage
	lock     sync.Mutex
}

func (m *Messages) Push(value float64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.received[int(value)] = struct{}{}
}

func (m *Messages) Exists(value float64) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.received[int(value)]
	return ok
}

func (m *Messages) Read() []int {
	m.lock.Lock()
	defer m.lock.Unlock()

	arr := make([]int, len(m.received))
	for key := range m.received {
		arr = append(arr, key)
	}

	return arr
}

func main() {
	n := maelstrom.NewNode()

	s := Server{n, Messages{
		received: make(map[int]struct{}),
		lock:     sync.Mutex{},
	},
	}

	s.n.Handle("broadcast", s.Broadcast)
	s.n.Handle("read", s.Read)
	s.n.Handle("topology", s.Topology)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
