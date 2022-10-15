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

## Join as a validator

### Requirements

Check out these [instructions](./passage-1/README.md#node-requirements) for installing `passage@v1.0.0`

If you haven't initialized your node, init passage chain by running

```sh
passage init --chain-id passage-1 <my_node_moniker>
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
    version: v1.0.0
    commit: 6ae7171e42f24203dc11369e7aef6d590bd09a47
    build_tags: netgo,ledger
    go: go version go1.17 linux/amd64
    ```

- Step-2: Download the mainnet genesis
    ```sh
    curl -s https://raw.githubusercontent.com/envadiv/mainnet/main/passage-1/genesis.json > ~/.passage/config/genesis.json
    ```

- Step-3: Verify genesis
    ```sh
    jq -S -c -M '' ~/.passage/config/genesis.json | shasum -a 256
    ```
    It should be equal to the contents in [checksum](passage-1/checksum.txt)

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
    ecfd6a2ab8dc2b196080ff6506cd0d1c68f6f8b5@passage-seed.panthea.eu:40656
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
  --chain-id=passage-1 \
  --commission-rate="0.05" \
  --commission-max-rate="0.10" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
  --from=<your_wallet_name>
```
