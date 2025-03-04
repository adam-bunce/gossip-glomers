set -e
go build -o ./cmd/broadcast/efficent-broadcast-2/efficent_broadcast_2_binary ./cmd/broadcast/efficent-broadcast-2/main.go
./maelstrom/maelstrom test -w broadcast --bin ./cmd/broadcast/efficent-broadcast-2/efficent_broadcast_2_binary --node-count 25 --time-limit 20 --rate 100 --latency 100

echo "$(sed -n "30p" < ./store/latest/results.edn | xargs) <= 20"
echo "$(grep --after-context 4 ":stable-latencies" ./store/latest/results.edn | xargs) (@0.5 <= 1s)"
echo "$(grep --after-context 4 ":stable-latencies" ./store/latest/results.edn | xargs) (@1 <= 2s)"


#:msgs-per-op 5.58029}, <= 20
#:stable-latencies {0 0, 0.5 678, 0.95 792, 0.99 890, 1 899}, (@0.5 <= 1000)
#:stable-latencies {0 0, 0.5 678, 0.95 792, 0.99 890, 1 899}, (@1 <= 2000)

