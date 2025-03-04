set -e
go build -o ./cmd/broadcast/single-node-broadcast/single_node_broadcast_binary ./cmd/broadcast/single-node-broadcast/main.go
./maelstrom/maelstrom test -w broadcast --bin ./cmd/broadcast/single-node-broadcast/single_node_broadcast_binary  --node-count 1 --time-limit 20 --rate 10
