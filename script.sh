go run . build-genesis passage-1
go run . export-claim-records ./passage-1/genesis.json claim_records.cs
go run . add-claim-records ./passage-1/vesting-accounts-genesis.json claim_records.csv
go run . migrate-accounts ~/Downloads/export-4088500.json claim-passage-genesis.json migrated.json
go run . migrate-genesis-state ~/Downloads/export-4088500.json migrated.json migrated1.json