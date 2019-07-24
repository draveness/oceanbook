package oceanbook

import (
	"context"
	"testing"

	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
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

type InsertOrderServer struct {
	grpc.ServerStream
	trades []*oceanbookpb.Trade
}

func NewTestInsertOrderServer() *InsertOrderServer {
	return &InsertOrderServer{
		trades: []*oceanbookpb.Trade{},
	}
}

func (x *InsertOrderServer) Send(t *oceanbookpb.Trade) error {
	x.trades = append(x.trades, t)

	return nil
}

func TestInsertOrder(t *testing.T) {
	stopCh := make(chan struct{})
	svc := NewService(stopCh)

	request := &oceanbookpb.NewOrderBookRequest{
		Pair: "BTC/CNY",
	}

	newOrderBookResponse, err := svc.NewOrderBook(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, &oceanbookpb.NewOrderBookResponse{}, newOrderBookResponse)

	stream := NewTestInsertOrderServer()
	err = svc.InsertOrder(&oceanbookpb.InsertOrderRequest{
		Id:       1,
		Price:    "1.0",
		Quantity: "2.0",
		Pair:     "BTC/CNY",
		Side:     oceanbookpb.Order_ASK,
	}, stream)
	assert.Nil(t, err)
	assert.Equal(t, []*oceanbookpb.Trade{}, stream.trades)

	err = svc.InsertOrder(&oceanbookpb.InsertOrderRequest{
		Id:       2,
		Price:    "2.0",
		Quantity: "1.0",
		Pair:     "BTC/CNY",
		Side:     oceanbookpb.Order_BID,
	}, stream)
	assert.Nil(t, err)
	assert.Equal(t, []*oceanbookpb.Trade{
		{
			Price:    "1",
			Quantity: "1",
			TakerId:  2,
			MakerId:  1,
		},
	}, stream.trades)
}

func TestCancelOrder(t *testing.T) {
	stopCh := make(chan struct{})
	svc := NewService(stopCh)

	cancelOrderResponse, err := svc.CancelOrder(context.Background(), &oceanbookpb.CancelOrderRequest{
		OrderId: 1,
		Pair:    "BTC/CNY",
	})
	assert.Equal(t, ErrOrderBookNotFound, err)
	assert.Nil(t, cancelOrderResponse)

	request := &oceanbookpb.NewOrderBookRequest{
		Pair: "BTC/CNY",
	}

	newOrderBookResponse, err := svc.NewOrderBook(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, &oceanbookpb.NewOrderBookResponse{}, newOrderBookResponse)

	stream := NewTestInsertOrderServer()
	err = svc.InsertOrder(&oceanbookpb.InsertOrderRequest{
		Id:       1,
		Price:    "1.0",
		Quantity: "2.0",
		Pair:     "BTC/CNY",
		Side:     oceanbookpb.Order_ASK,
	}, stream)
	assert.Nil(t, err)
	assert.Equal(t, []*oceanbookpb.Trade{}, stream.trades)

	cancelOrderResponse, err = svc.CancelOrder(context.Background(), &oceanbookpb.CancelOrderRequest{
		OrderId: 1,
		Pair:    "BTC/CNY",
	})
	assert.Nil(t, err)
	assert.Equal(t, &oceanbookpb.CancelOrderResponse{}, cancelOrderResponse)

	orderbook, _ := svc.orderbooks[request.Pair]
	assert.Equal(t, 0, orderbook.Bids.Size())
	assert.Equal(t, 0, orderbook.Asks.Size())
}
