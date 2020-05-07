![](https://github.com/hyperledger-labs/fabex/workflows/build/badge.svg) ![](https://github.com/hyperledger-labs/fabex/workflows/unit-tests/badge.svg) [![Total alerts](https://img.shields.io/lgtm/alerts/g/hyperledger-labs/fabex.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/hyperledger-labs/fabex/alerts/)


## **Block explorer for Hyperledger Fabric**

[Prerequisites](#prerequisites)

[Microservice mode](#microservice)

[CLI usage](#cli)



### <a name="prerequisites">**Prerequisites**</a>

1. Configure `config.yaml` (it's main config of the Fabex) and `connection-profile.yaml` (Hyperledger Fabric connection profile)

2. Start your Fabric blockchain network or sample [first network](https://github.com/hyperledger/fabric-samples/tree/release-1.4/first-network)

3. Install and start database (MongoDB or CassandraDB)
    
    If you choose Mongo:
      1. set initial user name and password in `db/mongo-compose/docker-compose.yaml`
      2. start container:
    
      ```
        cd db/mongo-compose
        docker-compose -f docker-compose.yaml up -d
      ```
    If you choose Cassandra:
      ``` 
      docker run --name cassandra --net=host -d cassandra
      ```

### <a name="microservice">**Microservice mode**</a>

You can start Fabex as standalone microservice:

    `./fabex -task=grpc -configpath=configs -configname=config -enrolluser=true -db=cassandra`
    
  or with Mongo
  
    `./fabex -task=grpc -configpath=configs -configname=config -enrolluser=true -db=mongo`

Use [fabex.proto](https://github.com/hyperledger-labs/fabex/blob/master/proto/fabex.proto) as service contract.

[Example](https://github.com/hyperledger-labs/fabex/blob/master/client/client.go) of GRPC client implementation.

   
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

---

You can choose database for data saving and retrieving with `-db flag` (MongoDB or CassandraDB):

    `./fabex -task=explore -configpath=configs -configname=config -db=mongo`
    `./fabex -task=explore -configpath=configs -configname=config -db=cassandra`


