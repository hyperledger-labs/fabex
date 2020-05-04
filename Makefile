start-fabric:
	@cd tests/fabcar && ./startFabric.sh

stop-fabric:
	@cd tests/fabcar && ./stopFabric.sh

start-fabex:
	@go run fabex.go -task=grpc -configpath=tests -configname=config -enrolluser=true -db=mongo

mongo:
	@cd db/mongo-compose && docker-compose -f docker-compose.yaml up -d

stop-mongo:
	@docker rm -f fabexmongo \
	&& sudo rm -rf /etc/mongoconfig \
    && sudo rm -rf /etc/mongodb \
    && sudo rm -rf credstore \
    && sudo rm -rf cryptostore
