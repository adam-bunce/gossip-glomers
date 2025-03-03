go build -o ./cmd/echo/echo_binary ./cmd/echo/main.go
./maelstrom/maelstrom test -w echo --bin ./cmd/echo/echo_binary --node-count 1 --time-limit 10
