#!/bin/bash

if [[ $1 == "mainnet" ]]
then
    export CURRENT_BLOCK=$(curl -s https://cosmos-rpc.polkachu.com/block | jq -r '.result.block.header.height')
    export TRUST_HEIGHT=$[$CURRENT_BLOCK-1000]
    export TRUST_BLOCK=$(curl -s https://cosmos-rpc.polkachu.com/block\?height\=$TRUST_HEIGHT)
    export TRUST_HASH=$(echo $TRUST_BLOCK | jq -r '.result.block_id.hash')

    echo "Gaia trust height: $TRUST_HEIGHT"
    echo "Gaia trust hash: $TRUST_HASH"
elif [[ $1 == "testnet" ]]
then
    export CURRENT_BLOCK=$(curl -s https://rpc.state-sync-01.theta-testnet.polypore.xyz/block | jq -r '.result.block.header.height')
    export TRUST_HEIGHT=$[$CURRENT_BLOCK-1000]
    export TRUST_BLOCK=$(curl -s https://rpc.state-sync-01.theta-testnet.polypore.xyz/block\?height\=$TRUST_HEIGHT)
    export TRUST_HASH=$(echo $TRUST_BLOCK | jq -r '.result.block_id.hash')

    echo "Gaia trust height: $TRUST_HEIGHT"
    echo "Gaia trust hash: $TRUST_HASH"
else
  echo "Expected argument: mainnet|testnet"
  exit 1;
fi
