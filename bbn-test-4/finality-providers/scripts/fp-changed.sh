#!/bin/bash -eu

# USAGE:
# ./fp-changed.sh

# returns

if ! git remote get-url bbnRepo &> /dev/null; then
  git remote add bbnRepo https://github.com/babylonchain/networks.git
fi
git fetch bbnRepo main --force

function fpChanged {
  git diff --diff-filter=AM bbnRepo/main --name-only **/finality-providers/registry/
}
