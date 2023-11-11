# Passage Relaunch Instructions 

**1. Please note that this genesis file has been updated after taking a state export of `passage-1` at height 4088500. Validators who were validating on `passage-1` have to REUSE the signing key(priv_validator-key.json) to validate on `passage-2`**

**2. Please remove the halt-height setting from your app config i.e., `~/.passage/config/app.toml`. Set the halt-height to `0` before starting your nodes**

## Node Requirements

### Minimum hardware requirements
- 16GB RAM
- 4 CPUs
- 400G SSD
- Ubuntu 20.04+ (Recommended)

### Software requirements
- Go 1.20+

#### Install Golang

```sh
sudo apt update
sudo apt install build-essential jq -y
wget https://dl.google.com/go/go1.20.5.linux-amd64.tar.gz
tar -xvf go1.20.5.linux-amd64.tar.gz
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
go version # should be go1.20.5
```

#### Setup Passage

**Clone the repo and install Passage3D**
```sh
mkdir -p $GOPATH/src/github.com/envadiv
cd $GOPATH/src/github.com/envadiv
git clone https://github.com/envadiv/Passage3D && cd Passage3D
git fetch
git checkout v2.0.0
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
version: v2.0.0
commit: a8148052a7aed65b45c453bf3accb5bafb496a11
build_tags: netgo ledger,
go: go version go1.20.6 linux/amd64
```

**Init Binary**
```sh
PASSAGE_MONIKER="Replace_AviaOne_by_your_name"
passage init $PASSAGE_MONIKER --chain-id passage-2
```

#### Download the updated genesis file and wipe the previous data

```sh
curl -s https://raw.githubusercontent.com/envadiv/mainnet/main/passage-2/genesis.json > ~/.passage/config/genesis.json
passage tendermint unsafe-reset-all --home ~/.passage
```

#### Verify genesis
```sh    
jq -S -c -M '' ~/.passage/config/genesis.json | shasum -a 256```
It should be equal to the contents in [checksum](checksum.txt)
```

#### Update seeds

SEEDS="ad9f93c38fafff854cdd65741df556d043dd6edb@5.161.71.7:26656,fbdcc82eeacc81f9ef7d77d22120f4567457c850@5.161.184.142:26656"
sed -i -e "s|^seeds *=.*|seeds = \"$SEEDS\"|" $HOME/.passage/config/config.toml

```
ad9f93c38fafff854cdd65741df556d043dd6edb@5.161.71.7:26656
fbdcc82eeacc81f9ef7d77d22120f4567457c850@5.161.184.142:26656
```

#### Add your wallet used with the chain passage-1

```
passage keys add NAME_WALLET --recover
```

#### Start the passage services
```
sudo systemctl restart passage
```

You can check the logs by running
```
journalctl -u passage -f
```

You should be able to see the following information in the logs:
```
......
5:34PM INF Starting Node service impl=Node
5:34PM INF Genesis time is in the future. Sleeping until then... genTime=2023-07-31T15:00:00Z
5:34PM INF Starting pprof server laddr=localhost:6060
```
