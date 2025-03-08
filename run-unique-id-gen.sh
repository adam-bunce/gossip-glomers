set -e


go build -o ./cmd/unique-ids/unique_ids_binary ./cmd/unique-ids/main.go
./maelstrom/maelstrom test -w unique-ids --bin ./cmd/unique-ids/unique_ids_binary --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

