package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
)

type Server struct {
	n *maelstrom.Node
	// map[key] -> [[offset, value], [offset+1, value]],
	// offset is unique to node/key so all start from 0 and increase by 1
	l          map[string][][]int
	lCommitted map[string]int
	lLock      sync.Mutex
}

var (
	offsetIdx = 0
	valueIdx  = 1
)

func (s *Server) Send(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	key := body["key"].(string)
	value := body["msg"].(float64)

	offset := s.addKV(key, int(value))

	return s.n.Reply(msg, map[string]any{"type": "send_ok", "offset": offset})
}

func (s *Server) addKV(key string, value int) int {
	s.lLock.Lock()
	defer s.lLock.Unlock()

	logValue, ok := s.l[key]
	highestOffset := 0
	if !ok {
		s.l[key] = [][]int{}
	} else {
		highestOffset = len(logValue)
	}

	newOffset := highestOffset + 1

	s.l[key] = append(s.l[key], []int{newOffset, value})

	return newOffset
}

func (s *Server) Poll(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	offsets := body["offsets"].(map[string]any)

	res := make(map[string][][]int)
	for k, v := range offsets {
		res[k] = s.getLogs(k, int(v.(float64)))
	}

	return s.n.Reply(msg, map[string]any{"type": "poll_ok", "msgs": res})
}

func (s *Server) getLogs(key string, offset int) [][]int {
	var res [][]int
	for _, v := range s.l[key] {
		if v[offsetIdx] >= offset {
			res = append(res, v)
		}
	}

	return res
}

func (s *Server) CommitOffsets(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	offsets := body["offsets"].(map[string]any)

	offsetsMap := make(map[string]int)
	for k, v := range offsets {
		offsetsMap[k] = int(v.(float64))
	}

	s.lCommitted = offsetsMap
	return s.n.Reply(msg, map[string]any{"type": "commit_offsets_ok"})
}

func (s *Server) ListCommittedOffsets(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	keys := body["keys"].([]any)
	offsets := make(map[string]int)
	for _, k := range keys {
		offsets[k.(string)] = s.lCommitted[k.(string)]
	}

	return s.n.Reply(msg, map[string]any{"type": "list_committed_offsets_ok", "offsets": offsets})
}

func InitServer() Server {
	n := maelstrom.NewNode()
	return Server{
		n:     n,
		l:     make(map[string][][]int),
		lLock: sync.Mutex{},
	}
}

func main() {
	s := InitServer()

	s.n.Handle("send", s.Send)
	s.n.Handle("poll", s.Poll)
	s.n.Handle("commit_offsets", s.CommitOffsets)
	s.n.Handle("list_committed_offsets", s.ListCommittedOffsets)

	err := s.n.Run()
	if err != nil {
		log.Fatal(err)
	}
}
