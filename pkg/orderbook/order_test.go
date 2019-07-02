package orderbook

import (
	"testing"
	"time"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type suiteMatchOrderTester struct {
	suite.Suite
}

func (s *suiteMatchOrderTester) TestMatchOrderNormalCase() {
	askOrder := &Order{
		ID:        1,
		Side:      OrderSideAsk,
		Price:     decimal.NewFromFloat(2.0),
		Quantity:  decimal.NewFromFloat(3.0),
		CreatedAt: time.Now(),
	}

	bidOrder := &Order{
		ID:        2,
		Side:      OrderSideBid,
		Price:     decimal.NewFromFloat(2.1),
		Quantity:  decimal.NewFromFloat(3.0),
		CreatedAt: time.Now(),
	}

	trade := askOrder.Match(bidOrder)

	s.Equal(trade, &Trade{
		Price:    askOrder.Price,
		Quantity: decimal.NewFromFloat(3.0),
		TakerID:  2,
		MakerID:  1,
	})
}

func (s *suiteMatchOrderTester) TestMatchOrderNoMatch() {
	askOrder := &Order{
		ID:        1,
		Side:      OrderSideAsk,
		Price:     decimal.NewFromFloat(3.0),
		Quantity:  decimal.NewFromFloat(3.0),
		CreatedAt: time.Now(),
	}

	bidOrder := &Order{
		ID:        2,
		Side:      OrderSideBid,
		Price:     decimal.NewFromFloat(2.1),
		Quantity:  decimal.NewFromFloat(3.0),
		CreatedAt: time.Now(),
	}

	trade := askOrder.Match(bidOrder)

	s.Nil(trade)
}

func TestMatchOrder(t *testing.T) {
	tester := new(suiteMatchOrderTester)
	suite.Run(t, tester)
}

type suiteOrderComparatorTester struct{ suite.Suite }

func (s *suiteOrderComparatorTester) TestBidOrderComparator() {
	b1 := Order{
		ID:        1,
		Side:      OrderSideBid,
		Price:     decimal.NewFromFloat(1.0),
		CreatedAt: time.Now(),
	}

	b2 := Order{
		ID:        2,
		Side:      OrderSideBid,
		Price:     decimal.NewFromFloat(1.0),
		CreatedAt: time.Now().Add(200 * time.Second),
	}

	b3 := Order{
		ID:        3,
		Side:      OrderSideBid,
		Price:     decimal.NewFromFloat(2.0),
		CreatedAt: time.Now().Add(300 * time.Second),
	}

	b4 := Order{
		ID:        4,
		Side:      OrderSideBid,
		Price:     decimal.NewFromFloat(0.5),
		CreatedAt: time.Now().Add(400 * time.Second),
	}

	tree := rbt.NewWith(OrderComparator)
	tree.Put(b1.Key(), b1)
	tree.Put(b2.Key(), b2)
	tree.Put(b3.Key(), b3)
	tree.Put(b4.Key(), b4)

	var orderValues []Order
	for _, value := range tree.Values() {
		orderValues = append(orderValues, value.(Order))
	}

	s.Equal([]Order{b4, b2, b1, b3}, orderValues)
}

func TestOrderComparator(t *testing.T) {
	tester := new(suiteOrderComparatorTester)
	suite.Run(t, tester)
}
