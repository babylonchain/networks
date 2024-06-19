#!/bin/bash -eu

# USAGE:
# ./fp-to-csv.sh

# exports all the fps from the registry to CSV with name, website and security contact

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
CSV_PATH="${CSV_PATH:-$CWD/fps.csv}"

fps=$(ls -d $CWD/../registry/*)

echo "name,website,contact" > $CSV_PATH
for filePathRegistryFP in ${fps}; do
  moniker=$(cat "$filePathRegistryFP" | jq -r '.description.moniker')
  website=$(cat "$filePathRegistryFP" | jq -r '.description.website')
  securityContact=$(cat "$filePathRegistryFP" | jq -r '.description.security_contact')

  echo "$moniker,$website,$securityContact" >> $CSV_PATH
done

