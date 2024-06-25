# Finality Provider Information Registry

__Due to fulfillment of the finality provider quotas,
registration will be turned off for the `bbn-test-4`
on Monday, 24 June 2024 EoD AoE.__

The `bbn-test-4` testnet will focus on the security of the staked Bitcoins by
testing the user's interaction with the BTC signet network. This will be a
lock-only network without a Babylon chain operating, meaning that the only
participants of this testnet will be finality providers and Bitcoin stakers.
This effectively means that for the next testnet, finality providers will only
be receiving Bitcoin Signet delegations and not have to vote for blocks.

Bitcoin holders that stake their Bitcoin can use Babylon's staking web
application to select the finality provider they want to delegate
their attestation of power to. They do so by including the finality provider's
BTC public key in the self-custodial Bitcoin Staking script.
Babylon will employ a Bitcoin indexer that collects all staking transactions
and extracts the finality provider BTC public keys that receive delegations
for display in the staking web application.
While the BTC public key is the only identifying information required
for a finality provider, it does not expose all the information that a
finality provider might want to share to attract more stake delegations.

The Babylon web application will additionally employ the finality provider
information registry in this repository to display additional information
such as the finality provider's moniker, website, and identity.
To protect this registry against spam, we require finality providers to submit
a deposit using the self-custodial Bitcoin staking to lock `0.1 signet BTC` for
one year. The deposit will be fully in the custody of the finality provider,
but not be counted as active stake and can be retrieved
after the deposit period expires.

An entry can be created in this registry by opening a pull
request containing:

1. Their identifying information combined with their BTC public key.
2. A signature of the information using the corresponding BTC private key.
3. A proof of submitting their deposit and the deposit having sufficient
   confirmations.

Finality providers can submit their information prior to or after the testnet
launch. To be included in the initial list that is displayed in the staking
web app, they have to submit the information prior to the launch.

The rest of the document explains the steps to create your finality provider's
keys and submitting the required information to be included in the registry.

## Versions

- [etosd](https://github.com/babylonchain/finality-provider/blob/ae30623a634450db81ce1755839754cc822bf5e5) - ae30623a634450db81ce1755839754cc822bf5e5
- [stakercli](https://github.com/babylonchain/btc-staker/blob/9be9838ca1124b64660dd1bdd57790bd7cc74e11) - 9be9838ca1124b64660dd1bdd57790bd7cc74e11

## 1. Create Finality Provider Keys

Finality Provider BTC key generation is covered by steps 1-3 from
[this guide](https://github.com/babylonchain/finality-provider/blob/ae30623a634450db81ce1755839754cc822bf5e5/docs/eots.md).
These steps describe how to set up the EOTS manager and generate the finality
provider keys using it. In this phase, finality providers should only use the
EOTS manager to generate their BTC keys and sign their finality provider
information (covered later in this guide). In later stages, finality providers
will be expected to operate a live version of the EOTS manager in order to
provide economic security to PoS chains.

At the end of these steps, your finality provider Bitcoin key pair will be
generated. Make sure that you store the key pair or the mnemonic you have
generated in a safe place, as it is going to be needed for your participation
on PoS security in the future stages of the Babylon testnet. Finality providers
that don't have access to their keys, will not be able to transition to later
stages.

⚠ Store the __mnemonic__ used for keys creation in a safe place.

## 2. Deposit self-lock BTC

Finality providers that want to register their information must make a deposit
of `0.1 signet BTC` using the self-custodial Bitcoin Staking script.
This is required to keep the finality provider information registry open,
but protect it from spam and entities that do not make a real commitment to the project.
The deposit will be locked for `52560` blocks (i.e. ~one year),
and will not be counted towards the active stake of the system.
Note that the deposit is still fully in the custody of the creator of the
transaction, but will only become unlocked after the deposit period expires.

__⚠ Warning!__
The deposit amount of `10000000` signet satoshi is the minimum amount required
for registration. Any deposit below this number will be considered an **invalid
registration**. There is no need to deposit more than `10000000`, but
such deposits with higher value will be considered as
a valid registrations.

The deposit is a Bitcoin transaction with an output containing the deposit value
committing to the Babylon Bitcoin Staking script.
A special set of values should be used for the deposit to be a valid one.
More specifically,
to create a valid deposit, you can follow the steps
[in this guide](https://github.com/babylonchain/btc-staker/blob/9be9838ca1124b64660dd1bdd57790bd7cc74e11/docs/create-phase1-staking.md),
with the following flags on the
`stakercli transaction create-phase1-staking-transaction` command:

- `--finality-provider-pk=<finality_provider_eots_pk>` The public key of your finality provider
previous generated.
- `--staker-pk` The public key of the account who funds the deposit transaction.
- `--staking-amount=10000000`, i.e. 0.1 signet BTC
- `--staking-time=52560`, i.e. ~1 year
- `--magic-bytes=62627434` `"bbt4"` as hex
- `--covenant-committee-pks=50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0`
  - This public key does not have a discrete logarithm therefore rendering the
    unbonding and slashing paths of the Bitcoin Staking script unusable.
    This makes the timelock path the only usable path as it only requires the
    staker's key.
- `--covenant-quorum=1`
- `--network=signet`

The difference between `--staker-pk` and `--finality-provider-pk`
is that the `--finality-provider-pk` flag specifies the public key of the
finality provider being registered, while the `--staker-pk` flag specifies
the Bitcoin public key that funds the deposit transaction.
It is recommended that the finality provider public key should be securely
stored and not hold any funds. Therefore, for creating the deposit
transaction, it is recommended that one funds the deposit transaction
with a separate public key (i.e. the `--staker-pk`) that holds the funds.

```shell
stakercli transaction create-phase1-staking-transaction \
  --staker-pk=<your_generated_staker_pub_key> \
  --finality-provider-pk=<your_finality_provider_eots_pk> \
  --staking-amount=10000000 --staking-time=52560 --magic-bytes=62627434 \
  --covenant-committee-pks=50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0 \
  --covenant-quorum=1 --network=signet

{
  "staking_tx_hex": "020000000002404b4c00000000002251207c2649dc890238fada228d52a4c25fcef82e1cf3d7f53895ca0fcfb15dd142bb0000000000000000496a476262743400b91ea4619bc7b3f93e5015976f52f666ae4eb5c98018a6c8e41424905fa8591fa89e7caf57360bc8b791df72abc3fb6d2ddc0e06e171c9f17c4ea1299e677565cd5000000000"
}
```

With the phase1 staking transaction you still need to fund with
`bitcoin-cli fundrawtransaction` and then sign it with your BTC key
`bitcoin-cli signrawtransactionwithwallet`.

After signing the transaction you should have it in
[hex format](https://github.com/babylonchain/btc-staker/blob/9be9838ca1124b64660dd1bdd57790bd7cc74e11/docs/create-phase1-staking.md#sign-transaction).
It is possible to verify if your transaction has the correct parameters before
submitting it to the Signet BTC ledger by using the following script:

```shell
$ export FP_BTC_PK=<your_finality_provider_eots_pk>
$ export SIGNED_TX=<your_signed_tx_hex>

$ ./bbn-test-4/finality-providers/scripts/fp-check-tx.sh

Provided transaction is valid staking transaction!
```

The success of the above command means that the signed transaction in hex format
is ready for propagation to the Bitcoin ledger.
This can happen in several ways:

- Through the [bitcoin-cli sendrawtransaction](https://github.com/babylonchain/btc-staker/blob/9be9838ca1124b64660dd1bdd57790bd7cc74e11/docs/create-phase1-staking.md#submit-transaction) command (recommended).
- [blockstream](https://blockstream.info/testnet/tx/push) Website to paste the
signed bitcoin transaction hex.
- [bitcoin-submittx](https://github.com/laanwj/bitcoin-submittx) Public github
repository that generates binary for P2P transaction submission
- [allthatnode](https://www.allthatnode.com/bitcoin.dsrv) Public RPC node

```shell
curl https://bitcoin-testnet-archive.allthatnode.com \
  --request POST \
  --header 'content-type: text/plain;' \
  --data '{"jsonrpc": "1.0", "id": "1", "method": "sendrawtransaction", "params": ["02000000000101ffa5874fdf64a535a4beae47ba0e66278b046baf7b3f3855dbf0413060aaeef90000000000fdffffff03404b4c00000000002251207c2649dc890238fada228d52a4c25fcef82e1cf3d7f53895ca0fcfb15dd142bb0000000000000000496a476262743400b91ea4619bc7b3f93e5015976f52f666ae4eb5c98018a6c8e41424905fa8591fa89e7caf57360bc8b791df72abc3fb6d2ddc0e06e171c9f17c4ea1299e677565cd50c876f7f70d0000001600141b9b57f4d4555e65ceb98c465c9580b0d6b0d0f60247304402200ae05daea3dc62ee7f2720c87705da28077ab19e420538eea5b92718271b4356022026c8367ac8bcd0b6d011842159cd525db672b234789a8d37725b247858c90a120121020721ef511b0faee2a487a346fdb96425d9dd7fa79210adbe7b47f0bcdc7e29de00000000"]}'

f22b9a1892df0e50977455b85b65324b079a9f230c5a9dede5ac711b9415d15b
```

Once the transaction is submited onchain it outputs an transaction hash.
Wait a few minutes and make sure that the transaction is included in the
blockchain by using the explorer
`https://live.blockcypher.com/btc-testnet/tx/<btc_staking_tx_hash>`.

> Make sure that the transaction has at least 6 confirmations block before
> creation of the pull request.

Keep the following information for inclusion in your finality provider
information. This will be used to prove that you are indeed the submitter of
the deposit.

```json
{
  ...
  "deposit": {
    "tx_hash": "f22b9a1892df0e50977455b85b65324b079a9f230c5a9dede5ac711b9415d15b",
    "signed_tx": "02000000000101ffa5874fdf64a535a4beae47ba0e66278b046baf7b3f3855dbf0413060aaeef90000000000fdffffff03404b4c00000000002251207c2649dc890238fada228d52a4c25fcef82e1cf3d7f53895ca0fcfb15dd142bb0000000000000000496a476262743400b91ea4619bc7b3f93e5015976f52f666ae4eb5c98018a6c8e41424905fa8591fa89e7caf57360bc8b791df72abc3fb6d2ddc0e06e171c9f17c4ea1299e677565cd50c876f7f70d0000001600141b9b57f4d4555e65ceb98c465c9580b0d6b0d0f60247304402200ae05daea3dc62ee7f2720c87705da28077ab19e420538eea5b92718271b4356022026c8367ac8bcd0b6d011842159cd525db672b234789a8d37725b247858c90a120121020721ef511b0faee2a487a346fdb96425d9dd7fa79210adbe7b47f0bcdc7e29de00000000"
  }
}
```

## 3. Create your Finality Provider information object

After forking the current repository,
navigate to the `finality-providers` directory and create a file under the
`finality-providers/registry/${nickname}.json` path.
`${nickname}`, corresponds to a unique human readable nickname your finality
provider can be identified with (e.g. your moniker). It should not contain
white spaces or unrecognizable characters.

Inside this file, store the following JSON information corresponding to your
finality provider.

```json
{
  "description": {
    "moniker": "<moniker>",
    "identity": "<identity>",
    "website": "<website>",
    "security_contact": "<security_contact>",
    "details": "<details>"
  },
  "btc_pk": "<eots_btc_pk>",
  "commission": "<commission_decimal>",
  "deposit": {
    "tx_hash": "tx_hash",
    "signed_tx": "signed_tx_hex"
  }
}
```

Properties descriptions:

- `moniker`: nickname of the finality provider.
- `identity`: optional identity signature (e.g. UPort or Keybase).
- `website`: optional website link.
- `security_contact`: required email for security contact.
- `details`: any other optional detail information.
- `btc_pk`: the btc pub key as hex.
- `commision`: the commission charged from btc stakers rewards.
Comission will be parsed as `sdk.Dec`:
  - `"1.00"` represents 100% commission.
  - `"0.10"` represents  10% commission.
  - `"0.01"` represents  01% commission.
- `deposit`: contains data for proof of locking.
  - `tx_hash`: The transaction hash of the deposit.
  - `signed_tx`: The funded signed locking transaction as hex.

## 4. Sign the Finality Provider information

Once you create your finality provider information, you need to prove that you
are indeed the owner of the Bitcoin Public Key contained within it. You can do
so, by signing the file with the corresponding Bitcoin Private Key of your
finality provider. This is another step of validation that
guarantees that the information provided by the finality provider was not tempered
and that the finality provider posseses the private key of that particular pub key.

To sign the file,
head back to the guide that you used to create your finality provider keys,
and more specifically the
[signing step](https://github.com/babylonchain/finality-provider/blob/ae30623a634450db81ce1755839754cc822bf5e5/docs/eots.md#33-sign-schnorr-signatures).
After signing the file, you should get this output:

```json
{
  "key_name": "my-key-name",
  "pub_key_hex": "c23e674f8fd2f28756a1536339646b84d40cf7205a8bb48bc6c6c68043964ab3",
  "signed_data_hash_hex": "b123ef5f69545cd07ad505c6d3b4931aa87b6adb361fb492275bb81374d98953",
  "schnorr_signature_hex": "b91fc06b30b78c0ca66a7e033184d89b61cd6ab572329b20f6052411ab83502effb5c9a1173ed69f20f6502a741eeb5105519bb3f67d37612bc2bcce411f8d72"
}
```

The signature is specified by the `schnorr_signature_hex` field of the output.
A file should be created under `./finality-providers/sigs` with the filename
being the same as the finality provider information stored under `./finality-providers/registry`
but with the `.sig` extension (e.g. `${nickname}.sig`).
The content of the file should be the plain value of the `schnorr_signature_hex` field.

__⚠ Warning!__
The signature was generated by reading the entire file data, not only the file
content. For proper verification, the exact file used for signing should
be submited in the pull request.

## 5. Create Pull Request

Submit your finality provider information under the `registry` directory and
your signature under the `sigs` directory. Both file names should have the
same name (e.g. `${nickname}`), but with `.json` and `.sig` extensions respectively.
__Make sure that you submit exactly the same file that you signed to ensure proper
verification__.

If you have installed all the binaries in your path, check out locally if your
finality provider is valid before creating the pull request:

```shell
$ ./bbn-test-4/finality-providers/scripts/verify-new-fp-offchain.sh && \
  ./bbn-test-4/finality-providers/scripts/verify-new-fp-onchain.sh

From https://github.com/babylonchain/networks
 * branch            main       -> FETCH_HEAD
verify bbn-test-4/finality-providers/registry/my_nickname.json
fp moniker: my great moniker
fp nickname: my_nickname
fp btcpk: a89e7caf57360bc8b791df72abc3fb6d2ddc0e06e171c9f17c4ea1299e677565
fp signature: 5e39939ccf68b8d30e134e132fe0e234b0840db3f380e17c57a0170c77235af3a555d8ea59eaacfaf43eaaa55d740549ee7f74cf844ed10dda2c81303006c348
eotsd verify signature
Verification is successful!
stakercli check transaction
Provided transaction is valid staking transaction!
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  2728  100  2728    0     0   4790      0 --:--:-- --:--:-- --:--:--  4785
BTC check transaction
✅ 'my_nickname' is a valid fp registration
```
