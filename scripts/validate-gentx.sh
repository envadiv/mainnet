#!/bin/sh
PASSAGE_HOME="/tmp/passage$(date +%s)"
RANDOM_KEY="randompassagevalidatorkey"
CHAIN_ID=passage-1
DENOM=upasg
MAXBOND=50000000000 # 50000PASSAGE

GENTX_FILE=$(find ./$CHAIN_ID/gentxs -iname "*.json")
LEN_GENTX=$(echo ${#GENTX_FILE})

# Gentx Start date
start="2021-03-31 15:00:00Z"
# Compute the seconds since epoch for start date
stTime=$(date --date="$start" +%s)

# Gentx End date
end="2021-04-06 15:00:00Z"
# Compute the seconds since epoch for end date
endTime=$(date --date="$end" +%s)

# Current date
current=$(date +%Y-%m-%d\ %H:%M:%S)
# Compute the seconds since epoch for current date
curTime=$(date --date="$current" +%s)

if [[ $curTime < $stTime ]]; then
    echo "start=$stTime:curent=$curTime:endTime=$endTime"
    echo "Gentx submission is not open yet. Please close the PR and raise a new PR after 31-March-2021 15:00:00"
    exit 0
else
    if [[ $curTime > $endTime ]]; then
        echo "start=$stTime:curent=$curTime:endTime=$endTime"
        echo "Gentx submission is closed"
        exit 0
    else
        echo "Gentx is now open"
        echo "start=$stTime:curent=$curTime:endTime=$endTime"
    fi
fi

if [ $LEN_GENTX -eq 0 ]; then
    echo "No new gentx file found."
else
    set -e

    echo "GentxFile::::"
    echo $GENTX_FILE

    echo "...........Init Regen.............."

    git clone https://github.com/envadiv/Passage3D
    cd Passage3D
    git checkout v1.0.0-rc1
    make build
    chmod +x ./build/passage

    ./build/passage keys add $RANDOM_KEY --keyring-backend test --home $PASSAGE_HOME

    ./build/passage init --chain-id $CHAIN_ID validator --home $PASSAGE_HOME

    echo "..........Fetching genesis......."
    rm -rf $PASSAGE_HOME/config/genesis.json
    curl -s https://raw.githubusercontent.com/envadiv/mainnet/main/$CHAIN_ID/genesis-prelaunch.json >$PASSAGE_HOME/config/genesis.json

    # this genesis time is different from original genesis time, just for validating gentx.
    sed -i '/genesis_time/c\   \"genesis_time\" : \"2021-03-29T00:00:00Z\",' $PASSAGE_HOME/config/genesis.json

    GENACC=$(cat ../$GENTX_FILE | sed -n 's|.*"delegator_address":"\([^"]*\)".*|\1|p')
    denomquery=$(jq -r '.body.messages[0].value.denom' ../$GENTX_FILE)
    amountquery=$(jq -r '.body.messages[0].value.amount' ../$GENTX_FILE)

    echo $GENACC
    echo $amountquery
    echo $denomquery

    # only allow $DENOM tokens to be bonded
    if [ $denomquery != $DENOM ]; then
        echo "invalid denomination"
        exit 1
    fi

    # limit the amount that can be bonded

    if [ $amountquery -gt $MAXBOND ]; then
        echo "bonded too much: $amountquery > $MAXBOND"
        exit 1
    fi

    ./build/passage add-genesis-account $RANDOM_KEY 100000000000000$DENOM --home $PASSAGE_HOME \
        --keyring-backend test

    ./build/passage gentx $RANDOM_KEY 90000000000000$DENOM --home $PASSAGE_HOME \
        --keyring-backend test --chain-id $CHAIN_ID

    cp ../$GENTX_FILE $PASSAGE_HOME/config/gentx/

    echo "..........Collecting gentxs......."
    ./build/passage collect-gentxs --home $PASSAGE_HOME
    sed -i '/persistent_peers =/c\persistent_peers = ""' $PASSAGE_HOME/config/config.toml

    ./build/passage validate-genesis --home $PASSAGE_HOME

    echo "..........Starting node......."
    ./build/passage start --home $PASSAGE_HOME &

    sleep 5s

    echo "...checking network status.."

    ./build/passage status --node http://localhost:26657

    echo "...Cleaning the stuff..."
    killall passage >/dev/null 2>&1
    rm -rf $PASSAGE_HOME >/dev/null 2>&1
fi
