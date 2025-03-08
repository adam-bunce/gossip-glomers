// https://fly.io/dist-sys/4/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
)

type Server struct {
	nodeId string
	n      *maelstrom.Node
	lock   sync.Mutex
	kv     *maelstrom.KV
}

func (s *Server) Read(msg maelstrom.Message) error {
	networkTotal := 0

	for _, node := range s.n.NodeIDs() {
		value, err := s.kv.ReadInt(context.Background(), node)
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				continue
			}
			return err
		}
		networkTotal += value
	}

	return s.n.Reply(msg, map[string]any{"type": "read_ok", "value": networkTotal})
}

func (s *Server) Add(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	delta, ok := body["delta"]
	if !ok {
		return fmt.Errorf("delta is not float64, got %T, %v", delta, delta)
	}

	deltaInt := int(delta.(float64))

	nodeId := s.n.ID()
	value := 0
	existingValue, err := s.kv.ReadInt(context.Background(), nodeId)
	if err == nil {
		value = existingValue
	} else if maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
		return err
	}

	err = s.kv.CompareAndSwap(context.Background(), nodeId, value, value+deltaInt, true)
	if err != nil {
		if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
			return s.Add(msg)
		}
		return err
	}

	return s.n.Reply(msg, map[string]any{"type": "add_ok"})
}

func InitServer() Server {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	return Server{
		n:  n,
		kv: kv,
	}
}

func main() {
	s := InitServer()

	s.n.Handle("read", s.Read)
	s.n.Handle("add", s.Add)

	err := s.n.Run()
	if err != nil {
		log.Fatal(err)
	}
}
