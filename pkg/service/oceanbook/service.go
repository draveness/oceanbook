package oceanbook

import (
	"context"
	"errors"
	"sync"

	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	"github.com/draveness/oceanbook/pkg/order"
	"github.com/draveness/oceanbook/pkg/orderbook"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrOrderBookNotFound returns when orderbook not found.
	ErrOrderBookNotFound = errors.New("orderbook not found")

	// ErrInvalidOrderPrice returns when order price is invalid.
	ErrInvalidOrderPrice = errors.New("invalid order price")

	// ErrInvalidOrderQuantity returns when order quantity is invalid.
	ErrInvalidOrderQuantity = errors.New("invalid order quantity")

	// ErrInvalidOrderSide returns when order side is invalid.
	ErrInvalidOrderSide = errors.New("invalid order side")
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

func (s *Service) getOrderBook(pair string) (*orderbook.OrderBook, bool) {
	s.RLock()
	defer s.RUnlock()

	orderbook, ok := s.orderbooks[pair]

	return orderbook, ok
}

// GetDepth .
func (s *Service) GetDepth(ctx context.Context, request *oceanbookpb.GetDepthRequest) (*oceanbookpb.Depth, error) {
	return nil, nil
}

// NewOrderBook .
func (s *Service) NewOrderBook(ctx context.Context, request *oceanbookpb.NewOrderBookRequest) (*oceanbookpb.NewOrderBookResponse, error) {
	_, exists := s.getOrderBook(request.Pair)
	if exists {
		return &oceanbookpb.NewOrderBookResponse{}, nil
	}

	s.Lock()
	defer s.Unlock()

	s.orderbooks[request.Pair] = orderbook.NewOrderBook(request.Pair)

	log.Infof("[oceanbook.liquidity] new order book with pair %s", request.Pair)

	return &oceanbookpb.NewOrderBookResponse{}, nil
}

// InsertOrder .
func (s *Service) InsertOrder(request *oceanbookpb.InsertOrderRequest, stream oceanbookpb.Oceanbook_InsertOrderServer) error {
	od, exists := s.getOrderBook(request.Pair)
	if !exists {
		return ErrOrderBookNotFound
	}

	price, err := decimal.NewFromString(request.Price)
	if err != nil {
		return ErrInvalidOrderPrice
	}

	quantity, err := decimal.NewFromString(request.Quantity)
	if err != nil {
		return ErrInvalidOrderQuantity
	}

	var side order.Side
	switch request.Side {
	case oceanbookpb.Order_ASK:
		side = order.SideAsk

	case oceanbookpb.Order_BID:
		side = order.SideBid

	default:
		return ErrInvalidOrderSide
	}

	trades := od.InsertOrder(&order.Order{
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
	od, exists := s.getOrderBook(request.Pair)
	if !exists {
		return nil, ErrOrderBookNotFound
	}

	od.CancelOrder(&order.Order{
		ID: request.OrderId,
	})

	return &oceanbookpb.CancelOrderResponse{}, nil
}
