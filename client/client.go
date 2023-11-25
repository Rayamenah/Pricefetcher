package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RaymondAmenah/pricefetcher-microservice/proto"
	"github.com/RaymondAmenah/pricefetcher-microservice/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpc client
func NewGRPCClient(remoteAddr string) (proto.PriceFetcherClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := proto.NewPriceFetcherClient(conn)

	return c, nil
}

type Client struct {
	endpoint string
}

func New(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

// JSON Transport
// The FetchPrice function is the function that will be called by the client.
// this client function return the JSON value
func (c *Client) FetchPrice(ctx context.Context, ticker string) (*types.PriceResponse, error) {

	endpoint := fmt.Sprintf("%s?ticker=%s", c.endpoint, ticker)

	//create a new get request url endpoint
	req, err := http.NewRequest("get", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// execute the request by using the http.DefaultClient.Do function. The result is a
	//http.Response and an error. We return the error if there is one.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service responded with a non ok status code")
	}

	//return the priceResponse
	priceResp := new(types.PriceResponse)
	if err := json.NewDecoder(resp.Body).Decode(priceResp); err != nil {
		return nil, err
	}

	return priceResp, nil
}
