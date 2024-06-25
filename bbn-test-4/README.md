# Bitcoin Staking Registry

The next testnet (`bbn-test-4`) will focus on the security of the
staked Bitcoins by only testing the user’s interaction with the
BTC signet network.
This will be a lock-only network without a Babylon chain operating,
meaning that the only participants of this testnet will be finality providers
and Bitcoin stakers.

This repository serves as an information registry for this system,
since there is no PoS chain to store system information.
Such information includes:
1. *[Finality Providers Information](./finality-providers)
   which contains additional information about
   finality providers participating in the system such as their monikers,
   committed commission, and other identifying information.
   Finality providers wishing to register more information about themselves to
   display on the testnet web interfaces, should submit their information
   there.*

   __Registration is turned off for the current testnet, you can still
   test the creation and check if everything worked properly with the
   [scripts](./finality-providers/scripts/).__
2. [System parameters](./parameters)
   which are versioned parameters denoting what constitutes
   valid stake that is accepted by the lock-only staking system.
   These parameters are timestamped on Bitcoin for easy verification.
3. [Integration](./integration)
   which contains information on wallet integration on the testnet system and
   operating a staking provider back-end.
4. [Covenant Committee](./covenant-committee)
   which contains information about the covenant emulation committee.
