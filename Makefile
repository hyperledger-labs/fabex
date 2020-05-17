fabric-test:
	@cd tests/fabcar && ./startFabric.sh

stop-fabric-test:
	@cd tests/fabcar && ./stopFabric.sh \
	&& rm -rf credstore \
    && rm -rf cryptostore

fabex-test:
	@cd tests/chaincode/fabcar/go && GOPROXY="https://goproxy.io" GOSUMDB=off go mod vendor
	@go run fabex.go -task=grpc -configpath=tests -configname=config -enrolluser=true -db=mongo

fabex-mongo:
	@go run fabex.go -task=grpc -configpath=configs -configname=config -enrolluser=true -db=mongo

fabex-cassandra:
	@go run fabex.go -task=grpc -configpath=configs -configname=config -enrolluser=true -db=cassandra

cassandra:
	@docker run -v /var/lib/cassandra:/var/lib/cassandra --name cassandra --net=host -d cassandra

mongo:
	@cd db/mongo-compose && docker-compose -f docker-compose.yaml up -d

mongo-test:
	@cd tests/db/mongo-compose && docker-compose -f docker-compose.yaml up -d

stop-mongo-test:
	@docker rm -f fabexmongo

unit-tests:
	@cd blockfetcher && go test -v

integration-tests:
	@cd client/fabexclient && go test -v
	@cd ./rest && go test -v