start-fabric:
	@cd tests/fabcar && ./startFabric.sh

stop-fabric:
	@cd tests/fabcar && ./stopFabric.sh

start-fabex:
	@go run fabex.go -task=grpc -configpath=tests -configname=config -enrolluser=true -db=mongo

mongo:
	@cd tests/db/mongo-compose && docker-compose -f docker-compose.yaml up -d

stop-mongo:
	@docker rm -f fabexmongo \
    && rm -rf credstore \
    && rm -rf cryptostore
