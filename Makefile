all: passage-1

.PHONY: passage-1

passage-1:
	go run . build-genesis passage-1
	mv -f passage-1/genesis.json passage-1/genesis-prelaunch.json
	bash -x ./scripts/gen-genesis.sh
