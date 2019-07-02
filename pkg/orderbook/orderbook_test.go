package orderbook

import (
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"io/ioutil"
	"testing"

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

		var trades []*Trade
		for _, o := range ode.Orders {
			rawResult := strings.Split(o, ",")
			var result []string
			for _, r := range rawResult {
				result = append(result, strings.TrimSpace(r))
			}

			var side OrderSide
			switch result[1] {
			case "ASK":
				side = OrderSideAsk
			case "BID":
				side = OrderSideBid
			}
			id, _ := strconv.Atoi(result[0])
			price, _ := decimal.NewFromString(result[2])
			quantity, _ := decimal.NewFromString(result[3])
			newOrder := &Order{
				ID:       uint64(id),
				Side:     side,
				Price:    price,
				Quantity: quantity,
			}

			newTrades := orderBook.InsertOrder(newOrder)
			if len(newTrades) > 0 {
				trades = append(trades, newTrades...)
			}
		}

		var expectedTrades []*Trade
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
			expectedTrades = append(expectedTrades, &Trade{
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

func TestOrderBook(t *testing.T) {
	tester := new(suiteOrderBookTester)
	suite.Run(t, tester)
}
