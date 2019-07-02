package oceanbook

import (
	"context"
	"errors"
	"sync"

	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	"github.com/draveness/oceanbook/pkg/orderbook"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var (
	ErrOrderBookNotFound    = errors.New("orderbook not found")
	ErrInvalidOrderPrice    = errors.New("invalid order price")
	ErrInvalidOrderQuantity = errors.New("invalid order quantity")
)

// Service represents oceanbook service.
type Service struct {
	sync.RWMutex
	orderbooks map[string]*orderbook.OrderBook
}

// NewService returns an oceanbook service.
func NewService(stopCh <-chan struct{}) *Service {
	return &Service{
		orderbooks: map[string]*orderbook.OrderBook{},
	}
}

// GetDepth .
func (s *Service) GetDepth(ctx context.Context, request *oceanbookpb.GetDepthRequest) (*oceanbookpb.Depth, error) {
	return nil, nil
}

// NewOrderBook .
func (s *Service) NewOrderBook(ctx context.Context, request *oceanbookpb.NewOrderBookRequest) (*oceanbookpb.NewOrderBookResponse, error) {
	s.Lock()
	defer s.Unlock()

	_, exists := s.orderbooks[request.Pair]
	if exists {
		log.Infof("[oceanbook.liquidity] order book %s already exists", request.Pair)
		return &oceanbookpb.NewOrderBookResponse{}, nil
	}

	od := orderbook.NewOrderBook(request.Pair)
	s.orderbooks[request.Pair] = od
	log.Infof("[oceanbook.liquidity] new order book with pair %s", request.Pair)

	return &oceanbookpb.NewOrderBookResponse{}, nil
}

// InsertOrder .
func (s *Service) InsertOrder(request *oceanbookpb.InsertOrderRequest, stream oceanbookpb.Oceanbook_InsertOrderServer) error {
	od, exists := s.orderbooks[request.Pair]
	if !exists {
		log.Infof("[oceanbook.liquidity/runInsertOrder] orderbook with pair %s not found", request.Pair)
		return ErrOrderBookNotFound
	}

	price, err := decimal.NewFromString(request.Price)
	if err != nil {
		log.Fatalf("[oceanbook.liquidity/runInsertOrder] order with invalid price %s", request.Price)
		return nil
	}

	quantity, err := decimal.NewFromString(request.Quantity)
	if err != nil {
		log.Fatalf("[oceanbook.liquidity] order with invalid quantity %s", request.Quantity)
		return nil
	}

	var side orderbook.OrderSide
	switch request.Side {
	case oceanbookpb.Order_ASK:
		side = orderbook.OrderSideAsk

	case oceanbookpb.Order_BID:
		side = orderbook.OrderSideBid

	default:
		log.Fatalf("[oceanbook.liquidity] order with invalid side %s", request.Side)

	}

	// FIXME: return trades
	trades := od.InsertOrder(&orderbook.Order{
		ID:       request.Id,
		Side:     side,
		Price:    price,
		Quantity: quantity,
	})

	for _, trade := range trades {
		stream.Send(trade.Serialize())
	}

	return nil
}

// CancelOrder .
func (s *Service) CancelOrder(ctx context.Context, request *oceanbookpb.CancelOrderRequest) (*oceanbookpb.CancelOrderResponse, error) {
	od, exists := s.orderbooks[request.Pair]
	if !exists {
		log.Infof("[oceanbook.liquidity/runCancelOrder] orderbook with pair %s not found", request.Pair)
		return nil, ErrOrderBookNotFound
	}

	od.CancelOrder(&orderbook.Order{
		ID: request.OrderId,
	})

	return nil, nil
}
