package main

import (
	// "context"
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/RaymondAmenah/pricefetcher-microservice/client"
	"github.com/RaymondAmenah/pricefetcher-microservice/proto"
)

func main() {
	var (
		jsonAddr = flag.String("json", ":3000", "service is running on json transport")
		grpcAddr = flag.String("grpc", ":4000", "service is running on grpc transport")
		svc      = loggingService{priceService{}}
		ctx      = context.Background()
	)

	flag.Parse()

	grpcClient, err := client.NewGRPCClient(":4000")
	if err != nil {
		log.Fatal(err)
	}

	go func() {

		time.Sleep(3 * time.Second)
		resp, err := grpcClient.FetchPrice(ctx, &proto.PriceRequest{Ticker: "SAI"})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", resp)

	}()

	go makeGRPCServerAndRun(*grpcAddr, svc)
	server := NewJSONAPIServer(*jsonAddr, svc)
	server.Run()

}
