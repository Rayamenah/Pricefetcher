package main

import (
	"context"
	"fmt"
)

type metricService struct {
	next PriceService
}

func NewMetricService(next PriceService) PriceService {
	return &metricService{
		next: next,
	}
}

func (s *metricService) FetchPrice(ctx context.Context, ticker string) (price float64, err error) {
	fmt.Println("pushing metrics to prometheus")
	//your metrics storage. Push to prometheus (guage, counters)
	return s.next.FetchPrice(ctx, ticker)
}
