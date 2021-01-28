![](https://github.com/hyperledger-labs/fabex/workflows/build/badge.svg) ![](https://github.com/hyperledger-labs/fabex/workflows/unit-tests/badge.svg) 

<p align="center">
<img src="https://github.com/hyperledger-labs/fabex/blob/2.x/fabex.png">
</p>

## **Block explorer for Hyperledger Fabric**
<br>

### Reference

[Prerequisites](#prerequisites)

[Microservice mode](#microservice)

[UI](#ui)

[CLI usage](#cli)

[Testing](#testing)

[Compatibility](#compatibility)

<br>

##### _FabEx is inspired by [ledgerfsck](https://github.com/C0rWin/ledgerfsck)_

<br>

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

### <a name="microservice">**Microservice mode**</a>

You can start Fabex as standalone microservice with Cassandra blocks storage:

    ./fabex -task=grpc -configpath=configs -configname=config -enrolluser=true -db=cassandra
    
  or with Mongo storage
  
    ./fabex -task=grpc -configpath=configs -configname=config -enrolluser=true -db=mongo

Use [fabex.proto](https://github.com/hyperledger-labs/fabex/blob/master/proto/fabex.proto) as service contract.

[Example](https://github.com/hyperledger-labs/fabex/blob/master/client/example/client.go) of GRPC client implementation.

<br><br>

### <a name="ui">**UI**</a>

UI is available on port 5252

![UI](https://github.com/hyperledger-labs/fabex/blob/2.x/ui.png)

<br><br>
 
### <a name="cli">**CLI**</a>
Build Fabex executable binary file:  

    go build

Enroll admin user:  

    ./fabex -enrolluser=true

Save blocks data to db:

    ./fabex -task=explore -configpath=configs -configname=config -db=cassandra
    

Also you can start service for fetching blocks in daemon mode: 
 
    ./fabex -task=explore -configpath=configs -configname=config -db=cassandra
    
    
Get transactions of specific block (chain operation):  

    ./fabex -task=getblock -blocknum=14 -configpath=configs -configname=config -db=cassandra

Get all transactions (db operation):  

    ./fabex -task=getall -configpath=configs -configname=config -db=cassandra


You can choose database for data saving and retrieving with `-db flag` (MongoDB or CassandraDB):

    ./fabex -task=explore -configpath=configs -configname=config -db=mongo
    ./fabex -task=explore -configpath=configs -configname=config -db=cassandra


<br><br>

### <a name="testing">**Testing**</a>

unit tests: `make unit-tests`

integration tests: `make integration-tests`

<br>

### <a name="compatibility">**Compatibility**</a>

FabEx is compatible with HLF 2.x.x. It has been decided not to support versions 1.х.х due to the fact that FabEx uses both native Fabric packages and SDK, which, unfortunately, conflict (protos conflicts). 
In addition, old versions of the Fabric don't have the necessary functionality (or it is in other packages / spread over several packages), which is available in newer versions. It is very difficult to do at first what is planned for version 2.x.x, and then search for the necessary functionality in the packages of the old Fabric and then search the compatible version of the SDK. 

If you are brave enough to support older versions, please send a pull request.