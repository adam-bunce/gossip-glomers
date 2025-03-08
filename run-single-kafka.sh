set -e
go build -o ./cmd/kafka-style-log/single_binary ./cmd/kafka-style-log/single/main.go
./maelstrom/maelstrom test -w kafka --bin ./cmd/kafka-style-log/single_binary --node-count 1 --concurrency 2n --time-limit 20 --rate 1000
#  Everything looks good! ヽ(‘ー`)ノ
