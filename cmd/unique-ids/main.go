// https://fly.io/dist-sys/2/
package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
)

func main() {
	n := maelstrom.NewNode()
	maximum := 0

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any

		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return err
		}

		body["type"] = "generate_ok"
		// Lamport timestamps (sort of, clients should keep track of maximum they've seen
		// and send it back to servers, if clients max is > than servers max server should
		//update its max to preserve causality)
		body["id"] = fmt.Sprintf("%s_%d", msg.Dest, maximum)
		maximum++

		return n.Reply(msg, body)
	})

	err := n.Run()
	if err != nil {
		log.Fatal(err)
	}
}
