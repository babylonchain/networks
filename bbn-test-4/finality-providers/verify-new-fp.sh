#!/bin/bash -eu

# USAGE:
# ./verify-new-fp.sh

# checks for modified files in **/bbn-test-4/finality-providers/** compared to main branch
# and validates all checks to make sure it is a valid finality provider registration

if ! command -v jq &> /dev/null; then
  echo "⚠️ jq command could not be found!"
  echo "Install it by checking https://stedolan.github.io/jq/download/"
  exit 1
fi

if ! command -v eotsd &> /dev/null; then
  echo "⚠️ eotsd command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/finality-provider.git"
  exit 1
fi

if ! command -v stakercli &> /dev/null; then
  echo "⚠️ stakercli command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/btc-staker.git"
  exit 1
fi

if ! git remote get-url bbnRepo &> /dev/null; then
  git remote add bbnRepo https://github.com/babylonchain/networks.git
fi
git fetch bbnRepo main

ALL_CHANGED_FILES=$(git diff --diff-filter=AM bbnRepo/main --name-only **/finality-providers/registry/)
if [ ${#ALL_CHANGED_FILES} -lt 1 ]; then
  echo "No new finality provider to register"
  exit 0
fi

for filePathRegistryFP in ${ALL_CHANGED_FILES}; do
  echo "verify" "$filePathRegistryFP"

  moniker=$(cat "$filePathRegistryFP" | jq -r '.description.moniker')
  echo "fp moniker:" $moniker
  if [ ${#moniker} -lt 3 ]; then
    echo $moniker "has less than 3 characteres"
    exit 1
  fi

  nickname=$(basename "$filePathRegistryFP" .json)
  echo "fp nickname:" $nickname

  btcPk=$(cat "$filePathRegistryFP" | jq -r '.btc_pk')
  echo "fp btcpk:" $btcPk

  baseDir=$(dirname $filePathRegistryFP)
  signature=$(cat "$baseDir/../sigs/$nickname.sig" | xargs)
  echo "fp signature:" $signature

  echo "eotsd verify signature"
  eotsd verify-schnorr-sig "$filePathRegistryFP" --btc-pk $btcPk --signature $signature
  echo

  signedTx=$(cat "$filePathRegistryFP" | jq -r '.deposit.signed_tx')
  echo "stakercli check transaction"
  stakercli transaction check-phase1-staking-transaction \
    --covenant-committee-pks 50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0 --covenant-quorum 1 \
    --magic-bytes 62627434 --network signet --staking-transaction $signedTx --finality-provider-pk $btcPk \
    --staking-time 52560 --min-staking-amount=10000000

  txHash=$(cat "$filePathRegistryFP" | jq -r '.deposit.tx_hash')

  onchainTx=$(curl https://signet.bitcoinexplorer.org/api/tx/$txHash | jq)
  confirmations=$(echo $onchainTx | jq -r '.confirmations')
  signedTxOnChain=$(echo $onchainTx | jq -r '.hex')

  echo "BTC check transaction"
  if [[ "$confirmations" -lt "6" ]]; then
    echo "The tx ${txHash} has ${confirmations} confirmations, it should have at least 6"
    exit 1
  fi

  if [[ $signedTx != $signedTxOnChain ]]; then
    echo "Signed tx in json ${signedTx} is different than signed tx on signet ${signedTxOnChain}"
    exit 1
  fi
  echo "✅ '${nickname}' is a valid fp registration"
done