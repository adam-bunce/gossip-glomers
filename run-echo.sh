go build -o echo_binary ./cmd/echo/main.go
./maelstrom/maelstrom test -w echo --bin echo_binary --node-count 1 --time-limit 10
