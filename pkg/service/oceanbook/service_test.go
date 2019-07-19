package oceanbook

import (
	"context"
	"testing"

	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	"github.com/stretchr/testify/assert"
)

func TestNewOrderBook(t *testing.T) {
	stopCh := make(chan struct{})
	svc := NewService(stopCh)

	request := &oceanbookpb.NewOrderBookRequest{
		Pair: "BTC/CNY",
	}

	response, err := svc.NewOrderBook(context.Background(), request)

	orderbook, ok := svc.orderbooks[request.Pair]
	assert.Nil(t, err)
	assert.Equal(t, &oceanbookpb.NewOrderBookResponse{}, response)
	assert.True(t, ok)
	assert.Equal(t, request.Pair, orderbook.Pair, "orderbook with pair %s exists", request.Pair)

	response, err = svc.NewOrderBook(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, &oceanbookpb.NewOrderBookResponse{}, response)
}
