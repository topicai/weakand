#Search multiple shards
## Overview
The idea behind multiple shards is to support search through large number of documents and utilize multiple machines to process one query parallely.

## Requirements
1. The system should be able to split reverted index into multiple shards
3. The system should be able to handle shard failure
5. The system should be able to add more shards while keeping online
6. The system should be able to balance size of reverted index among all shards

## Architecture
![Alt text](http://g.gravizo.com/g?
  digraph G {
    aize ="4,4";
    Root -> Leaf1;
    Root -> Leaf2;
    Root -> Leaf3;
    PersistentStorage [shape=box]
    Root -> PersistentStorage
    Leaf1 -> PersistentStorage
    Leaf2 -> PersistentStorage
    Leaf3 -> PersistentStorage
    Etcd [shape=box]
    Root -> Etcd
    Leaf1 -> Etcd
    Leaf2 -> Etcd
    Leaf3 -> Etcd
  }
)

## Specifications
### Sharding strategy
We plan to use Document based sharding. A comparison between doc based sharding and term based sharding([stolen from Jeff Dean](http://web.stanford.edu/class/cs276/Jeff-Dean-Stanford-CS276-April-2015.pdf)):

By doc: each shard has index for subset of docs 

* pro: each shard can process queries independently
* pro: easy to keep additional per-doc information 
* pro: network traffic (requests/responses) small 
* con: query has to be processed by each shard 
* con: O(K*N) disk seeks for K word query on N shards Ways of Index Partitioning  

By term: shard has subset of terms for all docs 

* pro: K word query => handled by at most K shards
* pro: O(K) disk seeks for K term query 
* con: much higher network bandwidth needed. data about each term for each matching doc must be collected in one place
* con: harder to have per-doc information

### Service Discovery

Our service discovery is simple: Root needs to know which leaf is up running and needs to be notified when one leaf dies away.  Etcd is used for service discovery. Once a leaf node is up, it registers itself into etcd and have a TTL attached to registered key-value pair. And leaf node will also start a goroutine to update the key-value pair periodically. Root node watches all registered key-value(actually a directory is being watched) and will be notified if a new leaf node is registered or a leaf node fails to update its key within TTL.

### Metadata(may not needed)
In current design our metadata only contains the mapping from leaf node to its docs’ docId. It is not necessary to have the reverse mapping(from docId to leaf). We plan to store metadata in Etcd. If we do checkpoint inverted index and forward index into some shared persistent storage, we might not need metadata at all since inverted index and forward index are good enough to handle leaf failure and recover.

### Persistent Storage
Persistent storage is used by leaf node to checkpoint inverted index(or forward index as well). Ideally the storage should be something like a shared file system. Amazon s3 is a candidate.