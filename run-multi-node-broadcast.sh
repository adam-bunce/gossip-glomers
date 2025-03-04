set -e
go build -o ./cmd/broadcast/multi-node-broadcast/multi_node_broadcast_binary ./cmd/broadcast/multi-node-broadcast/main.go
./maelstrom/maelstrom test -w broadcast --bin ./cmd/broadcast/multi-node-broadcast/multi_node_broadcast_binary  --node-count 5 --time-limit 20 --rate 10
