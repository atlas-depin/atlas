# x/storage

## Overview
The `x/storage` module is the lead indexer and orchestrator of the DePIN storage network.
Participating DePIN storage providers and all files maintained by the network are indexed by this module. Storage accounts and payments are also handled by this module using a subscription-style mechanism. Rewards derived from payments are distributed fairly amongst providers based on storage utilization.

TODO: summary of features, jot-notes

## Contents
1. Storage Representation
2. Storage Provider Orchestration
3. Storage Subscriptions
4. File Management
5. DePIN Security Strategy
6. Rewards Distribution & Slashing

## 1. Storage Representation
Storage is represented in units of `1 byte-second`. In simple terms, one unit of storage is equivalent to storing one byte of data for one second. A unit of storage is represented on-chain by a "**storage credit**". This approach directly links storage consumption (bytes) with duration (seconds), allowing for fair and granular billing, excellent traceability of storage consumption, efficient rewards distribution and/or slashing, and the ability to pool credits into one source of distribution rather than iterating through individual credit gauges associated with each and every payment.


## 2. Storage Provider Orchestration
TODO DEFINE
- Storage provider creation on-chain + object that is indexed
- Storage provider deletion
- How are stray file replicas with handled?


## 3. Storage Subscriptions
A user can purchase a subscription that reserves as little as 1GB of storage for as little was 1 day. The user may also purchase new subscriptions along with existing ones of different sizes and durations. Subscriptions can also be upgraded, both in duration and size. During the upload process (defined in the following section), the user has the option to specify which subscription to use; otherwise, the blockchain will select the last purchased active subscription with available storage.

### Subscription Object
```proto
message StorageSubscription {
  Timestamp start;        // start timestamp
  Timestamp end;          // end timestamp
  string status;          // subscription status: ACTIVE (default), PENDING, EXPIRED
  int64 spaceAvailable;   // total space reserved by the subscription
  int64 spaceUsed;        // total space actively used by the subscription
  string credits;         // string represented math.Int of available storage credits
}
```

### Subscription Purchase & Expansion
Storage subscriptions are priced in USD, and they are purchased with the equivalent amount of XXX tokens necessary. The value of XXX token is tracked by `x/oracle`. The amount of storage purchased is translated into storage credits which are maintained by `x/storage` in the indexed storage susbcription objects. Users may also purchase storage subscriptions on the behalf of another user.

Similarily to purchasing a new storage subscription, existing ones may be expanded in both quantity and time. The flexibility of scaling in accross dimensions is enabled by the multi-dimensional nature of storage credits (byte-seconds). Upon expansion, new storage credits are minted in the storage subscription, and the `end` or `spaceAvailable` parameters are updated accordingly. There is no storage subscription shrinking or refund policy.

### Subscription Statuses
There are three subscription statuses: `Active`, `Expired`, and `Pending`. An `Active` status is the default status, and is hopefully self-explanatory. An `Expired` status indicates that the subscription has expired (obviously). An expired subscription is persists in the kvstore for historical indexing purposes. The `Pending` status indicates that a user has purchased a storage subscription on the behalf of another user, and the recipient has yet to accept it. This part of the status mechanism exists to prevent spam.

### Subscription Expiry
A subscription expiry is detected when [TODO].

Upon expiry, all remaining storage credits are sent to a "dump". Storage credits in this dump are still associated with value in the storage module account, hence it is used as protocol-owned liquidity (POL).

### Future Development: On-Demand Subscription
Currently, subscriptions are based on a fixed duration, while storage credits exist to support effective rewards distribution, file replica management, and subscription expansion. In a future development, a new type of subscription that does not have an expiry could be added. In this case, as long as the user has storage credits remaining, they may use them on-demand. These storage credits would be billed at a premium rate since they offer immense flexibility and have no contributions to the POL. The premium would be routed to POL.

## 4. File Management
Files are indexed in a kvstore under the following following prefix schema within the storage module: `files/<creator_address>/<subscription_id>/<merkle_hash>/<start_block>`. The start block of the file is necessary in case the user chooses to upload several copies of the same file; however, this structure still prevents a user from uploading the same file twice in the same block.

### File Object
```proto
message File {
  string merkle;
  string owner;
  string subscription;
  int64 file_size;
  int64 start = 5;
  string[] providers;
  int64 replicas = 7;
}
```

### CR~~U~~D - Create, READ, ~~UPDATE~~, DELETE
When a file is created on-chain, it is commited to the kvstore. The `spaceUsed` of the corresponding storage subscription is also incremented by the file size. Storage credits are not withdrawn from the subscription until a storage provider attests to having the file.

A file's on-chain index entry can be queried. If no start block is provided, the query will fetch the first file uploaded of the specified merkle hash and address.

Files cannot be "updated" on-chain, because that changes the merkle hash of the file which is equivalent to deleting the file and re-creating it (which can be done).

When a file is deleted on-chain, the remaining storage credits until its expiry is computed and routed back to the subscription that posted the file.


## 5. DePIN Security Strategy

### Proof-of-Persistence

### Future Works: KZG Commitment Algorithm
TODO


## 6. Rewards Distribution & Slashing
TODO
