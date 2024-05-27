#!/bin/bash -eu

# USAGE:
# ./fp-changed.sh

# returns the finality provider modified files

function fpChanged {
  originUrl=$(git remote get-url origin)
  currBranch=$(git rev-parse --abbrev-ref HEAD)
  if [ "main" == "$currBranch" ] && [[ "$originUrl" =~ .*"babylonchain/networks.git".* ]]; then
    echo "Merge on main"
    git diff HEAD~1 --name-only **/finality-providers/registry/
  else
    if ! git remote get-url bbnRepo &> /dev/null; then
      git remote add bbnRepo https://github.com/babylonchain/networks.git
    fi
    git fetch bbnRepo main --force
    git diff --diff-filter=AM bbnRepo/main --name-only **/finality-providers/registry/
  fi
}
