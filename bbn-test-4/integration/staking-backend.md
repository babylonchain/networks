# Operating a Bitcoin Staking Backend

In this document we describe a reference tech stack employed by Babylon for a
back-end service that collects information about the testnet staking system from
Bitcoin and processes unbonding requests performed by users.
The entirety or part of this tech stack can be employed by staking providers
to build Bitcoin Staking applications.

This graphic demonstrates the architecture of the reference staking
back-end tech stack. In the following sections, we will go through the
components involved in more detail.
![Architecture](./assets/system-detailed.png)

## External Components

### Signet Bitcoin

Bitcoin serves as the decentralized ledger that stores and orders the staking
transactions for the lock-only test network. For the testnet,
the Bitcoin signet network is utilized.

The testnet system defines a set of
[global parameters](../parameters) that specify what constitutes a valid
staking transaction that the system recognizes. Transactions that are on
Bitcoin and adhere to the staking parameters are considered as valid staking
transactions.

### Staking Transactions

The system defines three types of transactions:
- *Staking* which creates new stake by locking signet BTC in the self-custodial
  Bitcoin Staking contract,
- *Unbonding* which unbonds stake from the Bitcoin staking contract before the locking expires, and
- *Withdraw* which extracts unlocked/unbonded stake to any wallet address specified by the staker.

The full spec of the staking transactions can be found
[here](https://github.com/babylonchain/babylon/blob/a8c9d27ab1d489eb55c23cbb2c75b87e1a85afdb/docs/staking-script.md).

Participants of the system can create staking transactions either through the
staking dApp or the staker CLI:
- The [Staking dApp](https://github.com/babylonchain/simple-staking/)
  is a user-friendly application allowing users to create, sign, and submit
  staking transactions to the Bitcoin ledger.
  It can either be built directly inside the Bitcoin wallet or can connect
  to a wallet. It communicates with a back-end service (described later) that
  has access to an unbonding pipeline in order to submit unbonding requests.
- The [Staker CLI](https://github.com/babylonchain/btc-staker)
  is a command line tool for power users that want full control
  on how their staking transaction is constructed. Stakers utilize the CLI to
  construct Bitcoin staking transactions and are responsible for signing them
  through a wallet of their choice and submitting them to the Bitcoin ledger or
  the Staking API service in the case of unbonding.

One could also build their own staking application and interact with the staking backend.

### Covenant Emulation Committee

The covenant emulation committee is a set of entities responsible for approving
on-demand unbonding requests. The members of the committee operate a
[covenant signer](https://github.com/babylonchain/covenant-signer)
program which involves a server that accepts requests that contain unbonding
requests that require the covenant emulator’s signature.
If the requests are valid, then the signature is returned in the response.
The covenant signer servers are contacted by a back-end unbonding pipeline
(described later) which collects their signatures before forwarding the fully
signed transaction to Bitcoin.
Note that the unbonding transaction requires the staker's signature, so the
covenant committee cannot consume the staking transaction
without the staker's approval.

The list of covenant emulation committee members is a global parameter, and the
signer servers are shared among all instances of the staking back-ends.

## Back-End Staking System Components

### [Staking Indexer](https://github.com/babylonchain/staking-indexer)

The staking indexer is a daemon that monitors the Bitcoin ledger for Bitcoin
Staking and Unbonding transactions. It consumes the
[system parameters](../parameters) to identify whether a staking transaction
is valid and its activity status. Valid staking transactions are stored in a
database and sent to RabbitMQ queues for further consumption by clients.

### [Staking API](https://github.com/babylonchain/staking-api-service)

The staking API is a service that provides information about the state
of the staking system and collects unbonding requests for further processing.

The API consumes staking events from the RabbitMQ queues the staking indexer
writes to. Based on those events, it calculates the status of delegations
(e.g. delegation has been successfully unbonded) and calculates information
about the state of the system (e.g. staker delegations, TVL, etc.) to provide
to consumers. It monitors for delegation expiration by consuming from the
RabbitMQ queue the expiry checker writes to.

In the case of an unbonding request, the API verifies it and stores it in
a database for further consumption by the unbonding pipeline.

### [Expiry Checker](https://github.com/babylonchain/staking-expiry-checker)

The staking expiry checker is a micro-service that reads staking transactions
from a database and checks whether their timelock has expired by comparing it
to the current Bitcoin height.
Once a delegation expires,
the expiry checker submits an event to a RabbitMQ queue for further consumption
by clients.

### [Unbonding Pipeline](https://github.com/babylonchain/cli-tools/)

The unbonding pipeline is a process that is run periodically to execute
pending unbonding requests.
It reads from the database the API writes unbonding requests to,
verifies the validity of the requests, and
connects to the covenant emulator committee members to collect their signatures.
Once a sufficient number of signatures is collected,
the fully signed unbonding transaction is sent to Bitcoin.

## Interoperability

Interoperability between staking providers requires that
there’s consensus on the following:
- *Staking Parameters*: All staking providers have utilize the same
  staking parameters.
- *Staking Transactions*: All staking providers validate the
  staking transactions in the same way and reach the same conclusions on their
  status (e.g. actie, expired, unbonding, etc.)
