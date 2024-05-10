# New Finality Provider

Follow the [README.md](../../bbn-test-4/finality-providers/README.md) instructions to generate information.

## Checklist

- [ ] [Create EOTS Key](https://github.com/babylonchain/finality-provider/blob/ae30623a634450db81ce1755839754cc822bf5e5/docs/eots.md?plain=1#L62)
- [ ] Backup Mnemonic
- [ ] [Make Deposit](../../bbn-test-4/finality-providers/README.md#2-deposit-self-lock-btc)
- [ ] Generate Finality Provider Information
- [ ] Input the information into a file under `bbn-test-4/finality-providers/registry/{nickname}.json`
- [ ] [Sign the file](https://github.com/babylonchain/finality-provider/blob/ae30623a634450db81ce1755839754cc822bf5e5/docs/eots.md?plain=1#L124)
- [ ] Input the signature into a file under `bbn-test-4/finality-providers/sigs/{nickname}.sig`

> [!CAUTION]
> The loss of the (generated keys + mnemonic) makes the finality provider useless
and unable to provide finality, which would lead to no transition to later phases
of the Babylon mainnet.
