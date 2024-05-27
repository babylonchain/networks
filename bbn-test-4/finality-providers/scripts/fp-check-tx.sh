#!/bin/bash -eu

# USAGE:
# ./fp-check-tx.sh

# check if the self-lock staking transaction has all the correct parameters set.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

STAKERCLI_BIN="${STAKERCLI_BIN:-stakercli}"
SIGNED_TX="${SIGNED_TX}"
FP_BTC_PK="${FP_BTC_PK}"

$STAKERCLI_BIN transaction check-phase1-staking-transaction \
  --covenant-committee-pks 50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0 --covenant-quorum 1 \
  --magic-bytes 62627434 --network signet --staking-transaction $SIGNED_TX --finality-provider-pk $FP_BTC_PK \
  --staking-time 52560 --min-staking-amount=10000000