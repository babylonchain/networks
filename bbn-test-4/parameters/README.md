# Staking Parameters

The staking parameters are governance parameters that specify what constitutes
a valid staking transaction that should be considered as an active one for
the lock-only testnet system.
They are maintained by Babylon and are timestamped on Bitcoin by a Bitcoin
governance wallet owned by it. They are additionally included in a GitHub
registry for easy retrieval and timestamp verification.

## Specification

The `global-params.json` file contains a JSON array (`versions`), with each
array element representing one version of the testnet parameters. The array
elements are ordered by increasing version.

```json
{
  "versions: [
    {
      "version": <params_version>,
      "activation_height": <bitcoin_activation_height>,
      "staking_cap": <satoshis_staking_cap_of_version>,
      "tag": "<magic_hex_encoded_bytes_to_identify_staking_txs>",
      "covenant_pks": [
        "<covenant_btc_pk1>",
        "<covenant_btc_pk2>",
        ...
      ],
      "covenant_quorum": <covenant_quorum>,
      "unbonding_time": <unbonding_time_btc_blocks>,
      "unbonding_fee": <unbonding_fee_satoshis>,
      "max_staking_amount": <max_staking_amount_satoshis>,
      "min_staking_amount": <min_staking_amount_satoshis>,
      "max_staking_time": <max_staking_time_btc_blocks>,
      "min_staking_time": <min_staking_time_btc_blocks>,
      "confirmation_depth": <confirmation_depth>
    },
    ...
  ]
}
```

The hash of each version of the parameters is further timestamped on Bitcoin by
a Babylon owned governance wallet to enable easy verification.  

A parameters version has the following rules:
- *Version*: The version should be an integer and versions should be
  monotonically increasing by `1` with an initial value of `0`.
- *ActivationHeight*: The activation height describes the Bitcoin height from
  which the parameters of this version are taken into account. Each new
  version, should have a strictly larger activation height than the previous
  version. This ensures that for any transaction, we can identify which staking
  parameters should apply to it.
- *StakingCap*: The staking cap describes the limit of Bitcoins that are
  accepted in total for this parameters version. It includes Bitcoins that have
  been accepted in prior versions. A later version should have a larger or
  equal staking cap than a prior version. The staking cap should be strictly
  larger than the maximum staking amount.
- *CovenantPKs*: Specifies the public keys of the covenant committee.
- *CovenantQuorum*: Specifies the quorum required by the covenant committee for
  unbonding transactions to be confirmed.
- *UnbondingFee*: Specifies the required fee that an unbonding transaction
  should have in satoshis. Can change arbitrarily between versions.
- *MinStakingAmount/MaxStakingAmount*: Specify the range of acceptable staking
  amounts in satoshis. Can change arbitrarily between versions. The maximum
  should be larger or equal to the minimum.
- *MinStakingTime/MaxStakingTime*: Specify the range of acceptable staking
  periods in BTC blocks. Can change arbitrarily between versions. The maximum
  should be larger or equal to the minimum. The maximum cannot be larger than
  65535.
- *ConfirmationDepth*: The number of confirmations required for transactions
  to be deep enough on the Bitcoin ledger so that their reversal is highly
  improbable. Inclusion of a transaction in a block means the confirmation depth
  for the transaction is `1`. More appended blocks further increment its
  confirmation depth.

Rules specification:
```
Let v_n and v_m be versions `n` and `m` respectively, with `m > n`.

In between versions:
- v_m.Version == v_n.Version + (m - n)
- v_m.ActivationHeight > v_n.ActivationHeight
- v_m.StakingCap >= v_n.StakingCap

For a particular version:
- len(v_m.Tag) == 4
- ValidBTCPks(v_m.CovenantPks)
- len(v_m.CovenantPks) > 0
- v_m.CovenantQuorum <= len(v_m.CovenantPks)
- v_m.StakingCap > v_m.MaxStakingAmount
- v_m.MaxStakingAmount >= v_m.MinStakingAmount
- v_m.MaxStakingTime >= v_m.MinStakingTime
- v_m.MaxStakingTime <= 65535
```
