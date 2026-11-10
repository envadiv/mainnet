# Passage3D Mainnet

This code is inspired from Regen Network's mainnet script

## Build genesis (For admin use)
### Address converter
Execute:
```shell
go run . addr-converter input.csv output_address.csv pasg 
```

### Building the `genesis.json` aka add all vesting, genesis accounts and balances

Execute:
```shell
go run . build-genesis passage-1
```

For pre-launch, we can ignore errors:

```shell
go run . build-genesis passage-prelaunch-1 --errors-as-warnings
```

### Add claim records/airdrop accounts
Execute:
```shell
go run . add-claim-records genesis.json claim-records.csv
```

Note: This will create a new genesis file from the input (`claim-passage-genesis.json`). Move to the network if you think it's final.
```shell
mv claim-passage-genesis.json <chain-id>/genesis.json
```

## passage-2 status

| | |
|---|---|
| Chain ID | `passage-2` |
| Running version | v4.0.3 ([release](https://github.com/envadiv/Passage3D/releases/tag/v4.0.3)) |
| Next upgrade | v4.1.0, CosmWasm security upgrade: release published, governance proposal not submitted yet ([instructions](passage-2/upgrades/v4.1.0.md)) |
| Minimum gas price | `12.5upasg` |

### Upgrade history

Heights from the chain (`passage q upgrade applied <name>`).

| Upgrade | Height | Instructions |
|---|---|---|
| v2.2.0 | 5288888 | [v2.2.0.md](passage-2/upgrades/v2.2.0.md) |
| v2.4.0 | 7080900 | [v2.4.0.md](passage-2/upgrades/v2.4.0.md) |
| v2.5.0 | 13331493 | [v2.5.0.md](passage-2/upgrades/v2.5.0.md) |
| v2.6.0 | 14273627 | [v2.6.0.md](passage-2/upgrades/v2.6.0.md) |
| v3.0.0 | 19807986 | [v3.0.0.md](passage-2/upgrades/v3.0.0.md) |
| v4.0.0 | 19950687 | [v4.0.0.md](passage-2/upgrades/v4.0.0.md) |
| v4.0.2 | 20114000 | [release](https://github.com/envadiv/Passage3D/releases/tag/v4.0.2) |
| v4.1.0 | to be announced | [v4.1.0.md](passage-2/upgrades/v4.1.0.md) |

v4.0.3 is a maintenance release on the v4.0.2 state machine; it needs no upgrade height.

## Join as a validator

### Requirements

Check out these [instructions](./passage-2/README.md#node-requirements) for installing `passage@v2.2.0`

If you haven't initialized your node, init passage chain by running

```sh
passage init --chain-id passage-2 <my_node_moniker>
```

### Start your validator node

- Step-1: Verify installation
    ```sh
    passage version --long
    ```

    it should display the following details:
    ```sh
    name: passage
    server_name: passage
    version: v2.2.0
    commit: 5377e66a3d7d2eafe7aad19299a6d9985ae63604
    build_tags: netgo,ledger
    go: go version go1.20.6 linux/amd64
    ```

- Step-2: Download the mainnet genesis
    ```sh
    curl -s https://raw.githubusercontent.com/envadiv/mainnet/main/passage-2/genesis.json > ~/.passage/config/genesis.json
    ```

- Step-3: Verify genesis
    ```sh
    jq -S -c -M '' ~/.passage/config/genesis.json | shasum -a 256
    ```
    It should be equal to the contents in [checksum](passage-2/checksum.txt)

- Step-4: Update seeds and persistent peers

    Open `~/.passage/config/config.toml` and update `persistent_peers` and `seeds` (comma separated list)
    #### Persistent peers
    ```sh
    69975e7afdf731a165e40449fcffc75167a084fc@104.131.169.70:26656,d35d652b6cb3bf7d6cb8d4bd7c036ea03e7be2ab@116.203.182.185:26656,ffacd3202ded6945fed12fa4fd715b1874985b8c@3.98.38.91:26656,8e0b0d4f80d0d2853f853fbd6a76390113f07d72@65.108.127.249:26656,0111da7144fd2e8ce0dfe17906ef6fd760325aca@142.132.213.231:26656
    ```

    #### Seeds
    ```
    aebb8431609cb126a977592446f5de252d8b7fa1@104.236.201.138:26656
    b6beabfb9309330944f44a1686742c2751748b83@5.161.47.163:26656
    7a9a36630523f54c1a0d56fc01e0e153fd11a53d@167.235.24.145:26656
    ```

- Step-5: Create systemd
    ```sh
    DAEMON_PATH=$(which passage)

    echo "[Unit]
    Description=passage daemon
    After=network-online.target
    [Service]
    User=${USER}
    ExecStart=${DAEMON_PATH} start
    Restart=always
    RestartSec=3
    LimitNOFILE=4096
    [Install]
    WantedBy=multi-user.target
    " >passage.service
    ```

- Step-6: Update system daemon and start passage node

    ```
    sudo mv passage.service /lib/systemd/system/passage.service
    sudo -S systemctl daemon-reload
    sudo -S systemctl enable passage
    sudo -S systemctl start passage
    ```

That's all! Your node should be up and running now. You can check the logs by running
```
journalctl -u passage -f
```

You would be able to see the following information in the logs:
```
......
5:34PM INF Starting Node service impl=Node
5:34PM INF Genesis time is in the future. Sleeping until then... genTime=2022-08-17T15:00:00Z
5:34PM INF Starting pprof server laddr=localhost:6060
```



You can query your node by executing the following command after the genesis time

```sh
passage status
```

### Create validator (Optional)
Note: This section is applicable for validators who wants to join post genesis time.

> **IMPORTANT:** Make sure your validator node is fully synced before running this command. Otherwise your validator will start missing blocks.

```sh
passage tx staking create-validator \
  --amount=9000000upasg \
  --pubkey=$(passage tendermint show-validator) \
  --moniker="<your_moniker>" \
  --chain-id=passage-2 \
  --commission-rate="0.05" \
  --commission-max-rate="0.10" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
  --from=<your_wallet_name>
```
