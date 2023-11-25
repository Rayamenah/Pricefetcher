build: 
	go build -o bin/pricefetcher-microservice 

run: build
	./bin/pricefetcher-microservice

proto: 
	 protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
   	 proto/service.proto

.PHONY: proto
 
