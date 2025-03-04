// https://fly.io/dist-sys/3e/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"slices"
	"sync"
	"time"
)

type MultiNodeBroadcaster interface {
	Broadcast(msg maelstrom.Message) error
	Read(msg maelstrom.Message) error
	Topology(msg maelstrom.Message) error
}

type Server struct {
	n *maelstrom.Node
	m Messages
	b Buffer
}

var mbnImplServer MultiNodeBroadcaster = &Server{}

func (s *Server) startBufferCheck() {
	go func() {
		for {
			select {
			case <-time.After(300 * time.Millisecond):
				s.sendBuffer()
			}
		}
	}()

}

func (s *Server) sendBuffer() {
	values := s.b.ReadAll()

	if len(values) > 0 {
		s.m.Push(values)
		body := map[string]any{"message": values, "type": "broadcast"}

		var destNodes []string
		if s.n.ID() == "n0" {
			for _, dest := range s.n.NodeIDs() {
				if dest == "n0" {
					continue
				}
				destNodes = append(destNodes, dest)
			}
		} else {
			destNodes = append(destNodes, "n0")
		}

		for _, dest := range destNodes {
			go func(dest string) {
				ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
				for {
					_, err := s.n.SyncRPC(ctx, dest, body)
					if err == nil {
						break
					}
				}
			}(dest)

		}
	}

}

func (s *Server) Broadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	var values []int

	switch v := body["message"].(type) {
	case float64:
		values = append(values, int(v))
	case []interface{}:
		for _, value := range v {
			switch v2 := value.(type) {
			case float64:
				values = append(values, int(v2))
			case int:
				values = append(values, v2)
			}
		}
	default:
		panic(fmt.Sprintf("server broadcast unhandled type %T : %v", v, v))
	}

	values = slices.DeleteFunc(values, s.m.Exists)
	if len(values) > 0 {
		s.b.Add(values)
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

func (m *Messages) Push(values []int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, value := range values {
		m.received[value] = struct{}{}
	}
}

func (m *Messages) Exists(value int) bool {
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

type Buffer struct {
	toSend []int
	lock   sync.Mutex
}

func (b *Buffer) Add(values []int) {
	b.lock.Lock()
	defer b.lock.Unlock()

	for _, value := range values {
		b.toSend = append(b.toSend, value)
	}
}

func (b *Buffer) ReadAll() []int {
	b.lock.Lock()
	defer b.lock.Unlock()

	values := make([]int, len(b.toSend))
	for i, value := range b.toSend {
		values[i] = value
	}

	return values
}

func main() {
	n := maelstrom.NewNode()

	s := Server{n, Messages{
		received: make(map[int]struct{}),
		lock:     sync.Mutex{},
	},
		Buffer{
			toSend: []int{},
			lock:   sync.Mutex{},
		},
	}

	s.startBufferCheck()
	s.n.Handle("broadcast", s.Broadcast)
	s.n.Handle("read", s.Read)
	s.n.Handle("topology", s.Topology)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
