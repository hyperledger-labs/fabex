<style>
#table {border: none;}
</style>

<table id="table"><tr>
<td><img src="https://github.com/hyperledger-labs/fabex/workflows/build/badge.svg" /></td>
<td><img src="https://github.com/hyperledger-labs/fabex/workflows/unit-tests/badge.svg" /></td>
<td>supported by OCRV</td><td><img align="right" width="70" height="70" src="https://avatars0.githubusercontent.com/u/60600953?s=100&v=4"></td></tr></table> 


## **Block explorer for Hyperledger Fabric**
<br>

### Reference

[Prerequisites](#prerequisites)

[Microservice mode](#microservice)

[CLI usage](#cli)

[Testing](#testing)

<br><br>

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

    `./fabex -task=grpc -configpath=configs -configname=config -enrolluser=true -db=cassandra`
    
  or with Mongo storage
  
    `./fabex -task=grpc -configpath=configs -configname=config -enrolluser=true -db=mongo`

Use [fabex.proto](https://github.com/hyperledger-labs/fabex/blob/master/proto/fabex.proto) as service contract.

[Example](https://github.com/hyperledger-labs/fabex/blob/master/client/example/client.go) of GRPC client implementation.

<br><br>
 
### <a name="cli">**CLI**</a>
Build Fabex executable binary file:  

    `go build`

Enroll admin user:  

    `./fabex -enrolluser=true`

Save blocks data to db:

    `./fabex -task=explore -configpath=configs -configname=config -db=cassandra`
    

Also you can start service for fetching blocks in daemon mode: 
 
    `./fabex -task=explore -configpath=configs -configname=config -db=cassandra -forever=true` 
    
    
Get transactions of specific block (chain operation):  

    `./fabex -task=getblock -blocknum=14 -configpath=configs -configname=config -db=cassandra`

Get all transactions (db operation):  

    `./fabex -task=getall -configpath=configs -configname=config -db=cassandra`


You can choose database for data saving and retrieving with `-db flag` (MongoDB or CassandraDB):

    `./fabex -task=explore -configpath=configs -configname=config -db=mongo`
    `./fabex -task=explore -configpath=configs -configname=config -db=cassandra`


<br><br>

### <a name="testing">**Testing**</a>

unit tests: `make unit-tests`

integration tests: `make integration-tests`