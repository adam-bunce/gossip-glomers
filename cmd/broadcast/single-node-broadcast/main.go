// https://fly.io/dist-sys/3a/
package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
)

type Messages struct {
	received []int
	lock     sync.Mutex
}

func (m *Messages) Push(value float64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.received = append(m.received, int(value))
}

func (m *Messages) Read() []int {
	m.lock.Lock()
	defer m.lock.Unlock()

	arr := make([]int, len(m.received))
	for _, val := range m.received {
		arr = append(arr, val)
	}

	return arr
}

func main() {
	n := maelstrom.NewNode()

	messages := Messages{
		received: []int{},
		lock:     sync.Mutex{},
	}

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		value, ok := body["message"].(float64)
		if !ok {
			log.Printf("field 'message' is not an int64, got %T: %v", body["message"], body["message"])
		}

		messages.Push(value)

		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{"type": "read_ok", "messages": messages.Read()})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
