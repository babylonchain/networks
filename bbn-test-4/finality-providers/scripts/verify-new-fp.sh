#!/bin/bash -eu

# USAGE:
# ./verify-new-fp.sh

# checks for modified files in **/bbn-test-4/finality-providers/** compared to main branch
# and validates the onchain and offchain verifications.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

$CWD/verify-new-fp-offchain.sh
$CWD/verify-new-fp-onchain.sh