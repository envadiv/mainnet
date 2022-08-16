# Passage3D Blockchain Mainnet Resrouces

## Resources

- ### [Block Explorers](./explorer-urls.txt)
- ### [Wallets](./wallets.txt)
- ### [API Nodes](./api-nodes.txt)
- ### [RPC Nodes](./rpc-nodes.txt)
- ### [Seed Nodes](./seed-nodes.txt)
- ### [Persistent Peer Nodes](./peer-nodes.txt)
- API Swagger Docs: https://api.passage.vitwit.com/swagger/

## Node Requirements

### Minimum hardware requirements
- 16GB RAM
- 4 CPUs
- 400G SSD
- Ubuntu 18.04+ (Recommended)

Note: 3 sentry architecture is the bare minimum setup required.

### Software requirements
- Go >=1.17.x & <=1.18.x

#### Install Golang

```sh
sudo apt update
sudo apt install build-essential jq -y
wget https://dl.google.com/go/go1.17.linux-amd64.tar.gz
tar -xvf go1.17.linux-amd64.tar.gz
sudo mv go /usr/local
```

```sh
echo "" >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin:$GOBIN' >> ~/.bashrc
```

Update PATH:
```sh
source ~/.bashrc
```

Verify Go installation:

```sh
go version # should be go1.17
```

#### Setup Passage

**Clone the repo and install Passage3D**
```sh
mkdir -p $GOPATH/src/github.com/envadiv
cd $GOPATH/src/github.com/envadiv
git clone https://github.com/envadiv/Passage3D && cd Passage3D
git fetch
git checkout v1.0.0
make install
```

**Verify installation**
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

## Gentx submission
This section applies to the validators who wants to join the genesis.

### Step-1: Initialize the chain
```sh
passage init --chain-id passage-1 <your_validator_moniker>
```

### Step-2: Replace the genesis
```sh
curl -s https://raw.githubusercontent.com/envadiv/mainnet/main/passage-1/genesis-prelaunch.json > $HOME/.passage/config/genesis.json
```

### Step-3: Add/Recover keys (Optional)
```sh
passage keys add <new_key>
```

or

```sh
passage keys add <key_name> --recover
```

### Step-4: Create Gentx
```sh
passage gentx <key_name> 9000000upasg  --chain-id passage-1
```

_Note: other amounts will be discarded_

ex:
```sh
passage gentx validator 9000000upasg --chain-id passage-1
```

**Note: Make sure to use the amount < available tokens in the genesis. Also max BONDED TOKENS allowed for gentxs are 9PASG or 9000000upasg**

You might be interested to specify other optional flags. For ex:

```sh
passage gentx validator 9000000upasg --chain-id passage-1 \
    --details <your validator details>
    --identity <The (optional) identity signature (ex. Keybase)>
    --commission-rate 0.05 \
    --commission-max-rate 0.2 \
    --commission-max-change-rate 0.01
```

It will show an output something similar to:
```
Genesis transaction written to "/home/ubuntu/.passage/config/gentx/gentx-8acdegf2783d0178781eefcf24f32a5e448e15a.json"
```

**Note: If you are generating gentx offline on your local machine, append `--pubkey` flag to the above command. You can get pubkey of your validator by running `passage tendermint show-validator`**

### Step-5: Fork envadiv mainnet repo
- Go to https://github.com/envadiv/mainnet
- Click on `fork` on the top right section of the page.

### Step-6: Clone mainnet repo
```sh
git clone https://github.com/<your_github_username>/mainnet $HOME/mainnet
```

### Step-7: Copy gentx to mainnet repo
```sh
cp ~/.passage/config/gentx/gentx-*.json $HOME/mainnet/passage-1/gentxs/
```

### Step-8: Commit and push to your repo
```sh
cd $HOME/mainnet
git add passage-1/gentxs/*
git commit -m "<your validator moniker> gentx"
git push origin master
```

### Step-9: Create gentx PR
- Go to your repository (on github)
- Click on Pull request and create a PR
- To make sure your submission is valid, please wait for the github action on your PR to complete
