set -e
go build -o ./cmd/grow-only-counter/grow_only_counter_binary ./cmd/grow-only-counter/main.go

./maelstrom/maelstrom test -w g-counter --bin ./cmd/grow-only-counter/grow_only_counter_binary --node-count 3 --rate 100 --time-limit 20 --nemesis partition

# :workload {:valid? true,
#            :errors nil,
#            :final-reads (1191 1191 1191),
#            :acceptable ([1191 1191])},
# :valid? true}
#
#
#Everything looks good! ヽ(‘ー`)ノ
