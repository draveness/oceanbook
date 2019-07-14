package orderbook

import (
	"math/rand"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"io/ioutil"
	"testing"

	"github.com/draveness/oceanbook/pkg/order"
	"github.com/draveness/oceanbook/pkg/trade"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type suiteOrderBookTester struct {
	suite.Suite
}

type OrderBookEntry struct {
	Name   string   `yaml:"name"`
	Orders []string `yaml:"orders"`
	Trades []string `yaml:"trades"`
}

func (ode *OrderBookEntry) Test(s *suiteOrderBookTester) {
	s.T().Run(ode.Name, func(t *testing.T) {
		orderBook := NewOrderBook("market")

		var trades []*trade.Trade
		for _, o := range ode.Orders {
			rawResult := strings.Split(o, ",")
			var result []string
			for _, r := range rawResult {
				result = append(result, strings.TrimSpace(r))
			}

			var side order.Side
			switch result[1] {
			case "ASK":
				side = order.SideAsk
			case "BID":
				side = order.SideBid
			}
			id, _ := strconv.Atoi(result[0])
			price, _ := decimal.NewFromString(result[2])
			quantity, _ := decimal.NewFromString(result[3])
			stopPrice := decimal.Zero

			if len(result) >= 5 {
				stopPrice, _ = decimal.NewFromString(result[4])
			}

			newOrder := &order.Order{
				ID:        uint64(id),
				Side:      side,
				Price:     price,
				Quantity:  quantity,
				StopPrice: stopPrice,
			}

			newTrades := orderBook.InsertOrder(newOrder)
			if len(newTrades) > 0 {
				trades = append(trades, newTrades...)
			}
		}

		var expectedTrades []*trade.Trade
		for _, t := range ode.Trades {
			rawResult := strings.Split(t, ",")
			var result []string
			for _, r := range rawResult {
				result = append(result, strings.TrimSpace(r))
			}

			price, _ := decimal.NewFromString(result[0])
			quantity, _ := decimal.NewFromString(result[1])
			makeID, _ := strconv.Atoi(result[2])
			takerID, _ := strconv.Atoi(result[3])
			expectedTrades = append(expectedTrades, &trade.Trade{
				Price:    price,
				Quantity: quantity,
				MakerID:  uint64(makeID),
				TakerID:  uint64(takerID),
			})
		}

		s.EqualValues(expectedTrades, trades)
	})
}

func (s *suiteOrderBookTester) TestInsertOrder() {
	orderbookFile, err := ioutil.ReadFile("./fixtures/orderbook.yaml")

	s.NoError(err)

	var entries []OrderBookEntry
	err = yaml.Unmarshal(orderbookFile, &entries)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		entry.Test(s)
	}
}

func (s *suiteOrderBookTester) TestInsertLimitOrder() {
	orderBook := NewOrderBook("market")

	limitOrder := &order.Order{
		ID:       2,
		Side:     order.SideBid,
		Price:    decimal.NewFromFloat(10.0),
		Quantity: decimal.NewFromFloat(30.0),
	}

	s.EqualValues([]*trade.Trade{}, orderBook.InsertOrder(limitOrder))
	s.EqualValues(limitOrder, orderBook.Bids.Right().Value.(*order.Order))
	s.EqualValues(1, orderBook.Bids.Size())
}

func (s *suiteOrderBookTester) TestInsertImmediateOrCancelOrder() {
	orderBook := NewOrderBook("market")

	iocOrder := &order.Order{
		ID:                2,
		Side:              order.SideBid,
		Price:             decimal.NewFromFloat(10.0),
		Quantity:          decimal.NewFromFloat(30.0),
		ImmediateOrCancel: true,
	}

	s.EqualValues([]*trade.Trade{}, orderBook.InsertOrder(iocOrder))
	s.True(orderBook.Bids.Empty())
	s.True(orderBook.Asks.Empty())
}

func (s *suiteOrderBookTester) TestCancelOrder() {
	orderBook := NewOrderBook("market")

	bidOrder := &order.Order{
		ID:       1,
		Side:     order.SideBid,
		Price:    decimal.NewFromFloat(10.0),
		Quantity: decimal.NewFromFloat(30.0),
	}

	askOrder := &order.Order{
		ID:       2,
		Side:     order.SideAsk,
		Price:    decimal.NewFromFloat(10.0),
		Quantity: decimal.NewFromFloat(30.0),
	}

	orderBook.InsertOrder(bidOrder)
	orderBook.InsertOrder(askOrder)

	orderBook.CancelOrder(bidOrder)
	s.Nil(orderBook.Bids.Right())
	s.EqualValues(0, orderBook.Bids.Size())

	orderBook.CancelOrder(askOrder)
	s.Nil(orderBook.Asks.Right())
	s.EqualValues(0, orderBook.Asks.Size())

	orderBook.InsertOrder(bidOrder)
	orderBook.CancelOrder(&order.Order{
		ID: 1,
	})
	s.Nil(orderBook.Bids.Right())
	s.EqualValues(0, orderBook.Bids.Size())
}

func TestOrderBook(t *testing.T) {
	tester := new(suiteOrderBookTester)
	suite.Run(t, tester)
}

func BenchmarkInsertOrder(b *testing.B) {
	orderBook := NewOrderBook("market")

	orders := make([]*order.Order, b.N)
	for n := 0; n < b.N; n++ {
		var side order.Side
		switch rand.Intn(2) {
		case 0:
			side = order.SideAsk
		case 1:
			side = order.SideBid
		}

		price := rand.Intn(10)
		quantity := rand.Intn(10) + 1

		orders[n] = &order.Order{
			ID:       uint64(n),
			Side:     side,
			Price:    decimal.NewFromFloat(float64(price)),
			Quantity: decimal.NewFromFloat(float64(quantity)),
		}
	}

	for n := 0; n < b.N; n++ {
		orderBook.InsertOrder(orders[n])
	}
}
