set -e
go build -o ./cmd/broadcast/fault-tolerant-broadcast/fault_tolerant_broadcast_binary ./cmd/broadcast/fault-tolerant-broadcast/main.go
./maelstrom/maelstrom test -w broadcast --bin ./cmd/broadcast/fault-tolerant-broadcast/fault_tolerant_broadcast_binary --node-count 5 --time-limit 20 --rate 10 --nemesis partition
