#!/bin/bash -eu

# USAGE:
# ./verify-new-fp-onchain.sh

# checks for modified files in **/bbn-test-4/finality-providers/** compared to main branch
# and validates if the onchain transaction deposit is valid.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

. $CWD/utils.sh
checkCommandJq

FP_CHANGED_FILES=$(fpChanged)
if [ ${#FP_CHANGED_FILES} -lt 1 ]; then
  echo "No new finality provider to register"
  exit 0
fi

for filePathRegistryFP in ${FP_CHANGED_FILES}; do
  echo "verify" "$filePathRegistryFP"

  txHash=$(cat "$filePathRegistryFP" | jq -r '.deposit.tx_hash')
  signedTx=$(cat "$filePathRegistryFP" | jq -r '.deposit.signed_tx')

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
  nickname=$(basename "$filePathRegistryFP" .json)
  echo "✅ '${nickname}' is a valid fp onchain-registration"
done