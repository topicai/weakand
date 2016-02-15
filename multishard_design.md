#Search multiple shards
## Overview
The idea behind multiple shards is to support search through large number of documents and utilize multiple machines to process one query parallely.

## Requirements
1. The system should be able to split reverted index into multiple shards
2. The system should be able to handle node failure
3. The system should be able to add more nodes while keeping online
4. The system should be able to balance size of reverted index among all nodes

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

### Metadata
Metadata would be stored in Etcd. I am thinking currently we only need to store the mapping from shard to node. For more information about what is shard in our context, see case study of elasticsearch.

### Persistent Storage
Persistent storage is used by leaf node to checkpoint inverted index(or forward index as well). Ideally the storage should be something like a shared file system. Amazon s3 is a candidate.

## Case Study
### Elastic Search(version 2.x current)
[Elasticsearch: The Definitive Guide](https://www.elastic.co/guide/en/elasticsearch/guide/current/index.html)

ElasticSearch is a distributed search engine based on Lucene. When talking about multiple shards, in ElasticSearch, a shard is a Lucene index. In our case, we can treat a shard as a collections of inverted index.
##### Route doc to shard
`shard = hash(DocId) % num_of_shards`

Questions: It is just a simple hash&mod function, how does Elasticsearch handles node failure? If a node fails, does it mean every doc has to recompute its shard and re-assign to a different shard?  
No, Elasticsearch has a fixed number of shards. Even one node fails, the number of shards keeps unchanged. **One node can be assigned with multipe shards.** In case of node failure, elsticsearch will copy the entire shard from failed node to other node.(Does it mean there is a shared filesystem here?)
 
##### Shard Mangement
Shard is the unit of scale in Elasticsearch. Since Elasticsearch has fixed numbe of shards, namely, shards can't be split or merged, it makes Elsticsearch a little diffcult to scale out of the box. User has to decide carefully how many shards to use in the initial phase. And once it reaches the capacity, user needs to do capacity planning, add more shards and reindex all the docs. This is a tradeoff made by the Elasticsearch team, quoted:
> Users often ask why Elasticsearch doesn’t support shard-splitting—the ability to split each shard into two or more pieces. The reason is that shard-splitting is a bad idea:
> 
1. Splitting a shard is almost equivalent to reindexing your data. It’s a much heavier process than just copying a shard from one node to another.
2. Splitting is exponential. You start with one shard, then split into two, and then four, eight, sixteen, and so on. Splitting doesn’t allow you to increase capacity by just 50%.
3. Shard splitting requires you to have enough capacity to hold a second copy of your index. Usually, by the time you realize that you need to scale out, you don’t have enough free space left to perform the split.
4. 
In a way, Elasticsearch does support shard splitting. You can always reindex your data to a new index with the appropriate number of shards (see Reindexing Your Data). It is still a more intensive process than moving shards around, and still requires enough free space to complete, but at least you can control the number of shards in the new index.

##### #Shards vs #Nodes
Ideally, each node should only have one shard. For each search query, the query is sent to all leaf nodes, and if one leaf node has multiple shards, it has to go through those shards sequentially. This somehow downgrades performance. In practice, `#shards == #nodes` may not always be true. But we should keep the number of shards in one node as minimal as possible.

##### Index Persistence
To easily recover from node failure(avoid reindex), we need to make index persistent to disk. Index persistence is essential for moving shard from failed node to other node. One naive implementation would be for each shard, we keep a big index in memory, and every so often, after some add/delete doc operations, we write the big index from memory to disk, replacing the old index that is written to disk last time. This approach requires encoding and writing index into disk every so often, since the index has the entire collection of docs for one shard, it will cause performance downgrade once the index becomes super large. This approach will also lose data if system crashes before index is persistent into disk.  
Elasticsearch solves this problem by using more than one index per shard. Instead of rewriting the whole inverted index, add new supplementary indices to reflect more-recent changes. Each inverted index can be queried in turn starting with the oldest and the results combined. And those indices are merged into large index on regular basis. Elasticsearch also introduces Transaction Log(like write ahead log in database) to record every operation. TransLog is flushed to disk for every operation so Elasicsearch can use it to recover if node crashes before index is persistent.  
For detailed explaination:[Elasticsearch Index Persistence](https://www.elastic.co/guide/en/elasticsearch/guide/current/making-text-searchable.html)

##### Primary and Replica shards
...




## Discussion
1.  Should we support split/merge of shards or just have a fixed number of shards like elasticsearch? I prefer a fixed number of shards as a start. It is easy to implement... 
2.  About index persistence. Just to start, shall we simply use the naive solution mentioned before(Keep a big index of the entire collections of docs in memory and flush this index to disk periodically)? And then we can move forward more advanced techniques. 

## More...
#### Service Discovery
what is the key-value format in etcd? Watcher, etc...
#### Distrubted Search Execution
rpc specified: retry, timeout, etc  
result merge  
rate limiting on root(reject requests if there are too many requests waiting for results)  
...
