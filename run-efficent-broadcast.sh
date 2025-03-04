set -e
go build -o ./cmd/broadcast/efficent-broadcast-1/efficent_broadcast_1_binary ./cmd/broadcast/efficent-broadcast-1/main.go
./maelstrom/maelstrom test -w broadcast --bin ./cmd/broadcast/efficent-broadcast-1/efficent_broadcast_1_binary --node-count 25 --time-limit 20 --rate 100 --latency 100

echo "$(sed -n "30p" < ./store/latest/results.edn | xargs) <= 30"
echo "$(grep --after-context 4 ":stable-latencies" ./store/latest/results.edn | xargs) (@0.5 <= 400)"
echo "$(grep --after-context 4 ":stable-latencies" ./store/latest/results.edn | xargs) (@1 <= 600)"

# :msgs-per-op 24.026756}, <= 30
# :stable-latencies {0 0, 0.5 177, 0.95 199, 0.99 203, 1 204}, (@0.5 <= 400)
# :stable-latencies {0 0, 0.5 177, 0.95 199, 0.99 203, 1 204}, (@1 <= 600)
# lowkey think stable latencies might just because my laptop is so fast yk
