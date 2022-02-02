module github.com/envadiv/mainnet

go 1.15

require (
	github.com/cockroachdb/apd/v2 v2.0.2
	github.com/cosmos/cosmos-sdk v0.45.0
	github.com/envadiv/Passage3D v1.0.0-rc1.0.20220131101710-543ab8ecab6f
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.14
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
