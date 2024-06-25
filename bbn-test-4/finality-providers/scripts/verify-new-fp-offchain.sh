#!/bin/bash -eu

# USAGE:
# ./verify-new-fp-offchain.sh

# checks for modified files in **/bbn-test-4/finality-providers/** compared to main branch
# and validates if the finality provider registration has valid values to send the
# transaction to the BTC chain.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

EOTSD_BIN="${EOTSD_BIN:-eotsd}"
STAKERCLI_BIN="${STAKERCLI_BIN:-stakercli}"

. $CWD/utils.sh
checkCommandJq

if ! command -v $EOTSD_BIN &> /dev/null; then
  echo "⚠️ $EOTSD_BIN command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/finality-provider/tree/37429abec0a514c4dbf95074fb231ff8464fdca8"
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

  securityContact=$(cat "$filePathRegistryFP" | jq -r '.description.security_contact')
  echo "fp security_contact:" $securityContact
  if [ ${#securityContact} -lt 4 ]; then
    echo $securityContact "has less than 4 characteres"
    exit 1
  fi

  commission=$(cat "$filePathRegistryFP" | jq -r '.commission')
  echo "fp commission:" $commission
  if ! [[ "$commission" =~ ^0\.[0-9]+$ ]]; then
    echo $commission "is not valid commision decimal, use 0.1 for 10%"
    exit 1
  fi

  nickname=$(basename "$filePathRegistryFP" .json)
  echo "fp nickname:" $nickname

  btcPk=$(cat "$filePathRegistryFP" | jq -r '.btc_pk')
  echo "fp btcpk:" $btcPk

  baseDir=$(dirname $filePathRegistryFP)
  signatureFilePath=$baseDir/../sigs/$nickname.sig
  if [ ! -f $signatureFilePath ]; then
    echo "signature file $signatureFilePath not found"
    exit 1
  fi

  signature=$(cat "$signatureFilePath" | xargs)
  echo "fp signature:" $signature

  echo "eotsd verify signature"
  $EOTSD_BIN verify-schnorr-sig "$filePathRegistryFP" --btc-pk $btcPk --signature $signature
  echo

  signedTx=$(cat "$filePathRegistryFP" | jq -r '.deposit.signed_tx')
  echo "stakercli check transaction"
  FP_BTC_PK=$btcPk SIGNED_TX=$signedTx STAKERCLI_BIN=$STAKERCLI_BIN $CWD/fp-check-tx.sh
  echo "✅ '${nickname}' is a valid fp offchain-registration"
done