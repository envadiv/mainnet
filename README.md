# Passage3D Mainnet

This code is inspired from Regen Network's mainnet script
## Address converter 
Execute:
```shell
go run . addr-converter input.csv output_address.csv pasg 
```

## Building genesis.json (For admin use)

Execute:
```shell
go run . build-genesis passage-1
```

For pre-launch, we can ignore errors:

```shell
go run . build-genesis passage-prelaunch-1 --errors-as-warnings
```

## Join as a validator

### Requirements

Check out these [instructions](./passage-1/README.md#Requirements) for installing `passage@v1.0.0`

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
    commit: [TBD]
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
    TBD
    ```
    #### Seeds
    ```sh
    TBD
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

That's all! Your node should be up and running now. You can query your node by executing the following command after the genesis time

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
