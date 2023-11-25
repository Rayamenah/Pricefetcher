package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// PriceFetcher is an interface that can fetch a price
type PriceService interface {
	FetchPrice(context.Context, string) (float64, error)
}

// priceFetcher implements the PriceFetcher interface
type priceService struct{}

var priceMocks = map[string]float64{
	"BTC": 20_000.0,
	"ETH": 2_000.0,
	"SAI": 10_000.0,
}

func (s *priceService) FetchPrice(_ context.Context, ticker string) (float64, error) {
	//mimic the HTTP process
	time.Sleep(100 * time.Millisecond)
	price, ok := priceMocks[ticker]
	if !ok {
		return price, fmt.Errorf("the given ticker (%s) is not supported", ticker)
	}
	return price, nil
}

type loggingService struct {
	priceService
}

// logging function for logging each time a price is fetched
// each time this function is called it logs immediately before the price is fetched
func (s loggingService) FetchPrice(ctx context.Context, ticker string) (price float64, err error) {
	defer func(begin time.Time) {
		logrus.WithFields(logrus.Fields{
			"requestID": ctx.Value("requestID"),
			"took":      time.Since(begin),
			"err":       err,
			"price":     price,
		}).Info("fetchPrice")
	}(time.Now())

	//return the price fetched after logging
	return s.priceService.FetchPrice(ctx, ticker)
}
