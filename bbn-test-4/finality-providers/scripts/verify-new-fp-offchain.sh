#!/bin/bash -eu

# USAGE:
# ./verify-new-fp-offchain.sh

# checks for modified files in **/bbn-test-4/finality-providers/** compared to main branch
# and validates if the finality provider registration has valid values to send the
# transaction to BTC chain.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
. $CWD/fp-changed.sh

EOTSD_BIN="${EOTSD_BIN:-eotsd}"
STAKERCLI_BIN="${STAKERCLI_BIN:-stakercli}"

if ! command -v jq &> /dev/null; then
  echo "⚠️ jq command could not be found!"
  echo "Install it by checking https://stedolan.github.io/jq/download/"
  exit 1
fi

if ! command -v $EOTSD_BIN &> /dev/null; then
  echo "⚠️ $EOTSD_BIN command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/finality-provider.git"
  exit 1
fi

if ! command -v $STAKERCLI_BIN &> /dev/null; then
  echo "⚠️ $STAKERCLI_BIN command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/btc-staker.git"
  exit 1
fi

FP_CHANGED_FILES=$(fpChanged)
if [ ${#FP_CHANGED_FILES} -lt 1 ]; then
  echo "No new finality provider to register"
  exit 0
fi

for filePathRegistryFP in ${FP_CHANGED_FILES}; do
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
  $EOTSD_BIN verify-schnorr-sig "$filePathRegistryFP" --btc-pk $btcPk --signature $signature
  echo

  signedTx=$(cat "$filePathRegistryFP" | jq -r '.deposit.signed_tx')
  echo "stakercli check transaction"
  $STAKERCLI_BIN transaction check-phase1-staking-transaction \
    --covenant-committee-pks 50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0 --covenant-quorum 1 \
    --magic-bytes 62627434 --network signet --staking-transaction $signedTx --finality-provider-pk $btcPk \
    --staking-time 52560 --min-staking-amount=10000000

  echo "✅ '${nickname}' is a valid fp offchain-registration"
done