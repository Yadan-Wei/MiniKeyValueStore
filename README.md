# Mini Key-Value Cache

## Introduction

This application is a distributed key-value cache designed with resilience and distributed consensus in mind, utilizing the robust HashiCorp Raft library. Key features include:

- **Key-Value Operations**:

  It enables clients to perform read and write operations on key-value pairs. Clients can set a value for a key or retrieve the value associated with a key.

- **Fault Tolerance**:

  The system is architected to withstand failures of individual nodes without disrupting the overall availability or consistency of the data.

- **Handling Network Partitions**:

  The application is capable of handling network partitions, ensuring that once the network is healed, the system converges to a consistent state.

- **Data Replication**:

  The key-value pairs are replicated across multiple nodes in the system, safeguarding against data loss and enabling high availability.

- **Snapshotting**:

  The system supports snapshotting of the current state, allowing for faster recovery and simplifying the process of bringing new nodes up to date.

- **Dynamic Cluster Membership**:

  Nodes can be added or removed from the cluster dynamically, allowing the system to scale with the demands of the workload.

This distributed mini key-value cache exemplifies a system that is resilient to the unpredictable nature of distributed environments, providing a reliable and consistent caching solution.

## Get Started

### Initialization

- Build Binary

In the root directory build the execution binary.

`go build -o minikv ./`

- Create the Leader Node

`./minikv --http=127.0.0.1:6000 --raft=127.0.0.1:7000 --node=1 --bootstrap=true `

- Create the First Follower Node

`./minikv --http=127.0.0.1:6001 --raft=127.0.0.1:7001 --node=2 --join=127.0.0.1:6000`

- Create the Second Follower Node

`./minikv --http=127.0.0.1:6002 --raft=127.0.0.1:7002 --node=3 --join=127.0.0.1:6000`

### Operations

- Write to the Leader

`curl "http://localhost:6000/set?key=test1&value=writetoleader"`

`curl "http://localhost:6000/set?key=test2&value=readfromfollower"`

- Read from the Leader

`curl "http://localhost:6000/get?key=test1"`

- Read from Followers

`curl "http://localhost:6001/get?key=test2"`

`curl "http://localhost:6002/get?key=test2"`

- Cluster Snapshot Recovery

Ctrl + C kill all nodes and restore all of them

`./minikv --http=127.0.0.1:6000 --raft=127.0.0.1:7000 --node=1`

`./minikv --http=127.0.0.1:6001 --raft=127.0.0.1:7001 --node=2`

`./minikv --http=127.0.0.1:6002 --raft=127.0.0.1:7002 --node=3`

`curl "http://localhost:6002/get?key=test2"`

`curl "http://localhost:6000/set?key=test3&value=clusteRecovery"`

`curl "http://localhost:6002/get?key=test3"`

- Leader Fault Tolerance

Kill the leader and do the write operation to new leader.

`curl "http://localhost:6001/set?key=test4&value=CheckLeaderSwitch"`

Restart the original leader and write to new leader.

`./minikv --http=127.0.0.1:6000 --raft=127.0.0.1:7000 --node=1`

Original leader becomes candidate and then follower.

`curl "http://localhost:6001/set?key=test5&value=CheckLeaderSwitchBackOrNot"`

- Follower Fault Tolerance

Kill one follower and restart it.

`./minikv --http=127.0.0.1:6001 --raft=127.0.0.1:7001 --node=2`

Leader won't change. Node becomes candiate and then follower.
