![](https://github.com/hyperledger-labs/fabex/workflows/build/badge.svg) ![](https://github.com/hyperledger-labs/fabex/workflows/unit-tests/badge.svg) 

<p align="center">
<img src="https://github.com/hyperledger-labs/fabex/blob/2.x/fabex.png">
</p>

## **Block explorer for Hyperledger Fabric**
<br>

### Reference

[Tutorial](#tutorial)

[Prerequisites](#prerequisites)

[Start fabex service](#start)

[UI](#ui)

[Testing](#testing)

<br>

##### _FabEx is inspired by [ledgerfsck](https://github.com/C0rWin/ledgerfsck)_

<br>

### <a name="tutorial">**Tutorial**</a>
the tutorial is available at [this link](https://medium.com/@vadiminshakov/fabex-tutorial-an-introduction-to-the-right-hyperledger-fabric-explorer-cd9ee1848cd9).

### <a name="prerequisites">**Prerequisites**</a>

1. Configure `config.yaml` (it's main config of the Fabex) and `connection-profile.yaml` (Hyperledger Fabric connection profile)

2. Install and start database (MongoDB or CassandraDB)
    
    If you choose Mongo:
      1. set initial user name and password in `db/mongo-compose/docker-compose.yaml`
      2. start container:
    
      ```
      make mongo
      ```
    If you choose Cassandra:
      ``` 
      make cassandra
      ```
3. (OPTIONAL) Start your Fabric blockchain network or sample test network with 
   ```
   make fabric-test
   ```

<br><br>

### <a name="start">**Start fabex service**</a>

You can start Fabex as standalone microservice with Cassandra blocks storage:

    CONFIG=config/config.yaml DB=cassandra ./fabex
    
  or with Mongo storage
  
    CONFIG=config/config.yaml DB=mongo ./fabex

Use [fabex.proto](https://github.com/hyperledger-labs/fabex/blob/master/proto/fabex.proto) as service contract.

[Example](https://github.com/hyperledger-labs/fabex/blob/master/client/example/client.go) of GRPC client implementation.

<br><br>

### <a name="ui">**UI**</a>

UI is available on port 5252

![UI](https://github.com/hyperledger-labs/fabex/blob/2.x/ui.png)

<br><br>

### <a name="testing">**Testing**</a>

unit tests: `make unit-tests`

integration tests: `make integration-tests`