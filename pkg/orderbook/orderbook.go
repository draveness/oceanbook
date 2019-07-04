package orderbook

import (
	"sync"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
	log "github.com/sirupsen/logrus"
)

// OrderBook is the order book.
type OrderBook struct {
	sync.RWMutex
	Pair string
	Bids *rbt.Tree
	Asks *rbt.Tree
}

// NewOrderBook returns a pointer to an orderbook.
func NewOrderBook(pair string) *OrderBook {
	return &OrderBook{
		Pair: pair,
		Bids: rbt.NewWith(OrderComparator),
		Asks: rbt.NewWith(OrderComparator),
	}
}

// InsertOrder inserts new order into orderbook.
func (od *OrderBook) InsertOrder(newOrder *Order) []*Trade {
	log.Debugf("[oceanbook.orderbook] insert order with id %od %s * %s, side %s", newOrder.ID, newOrder.Price, newOrder.Quantity, newOrder.Side)

	// TODO: deal with order with same id but different properties
	var takerBooks *rbt.Tree
	var makerBooks *rbt.Tree
	switch newOrder.Side {
	case OrderSideAsk:
		takerBooks = od.Asks
		makerBooks = od.Bids

	case OrderSideBid:
		takerBooks = od.Bids
		makerBooks = od.Asks

	default:
		log.Fatalf("[oceanbook.orderbook] invalid order side %s", newOrder.Side)
	}

	trades := []*Trade{}

	od.Lock()
	defer od.Unlock()

	_, found := takerBooks.Get(newOrder.Key())
	if found {
		return trades
	}

	for {
		best := makerBooks.Right()
		if best == nil {
			break
		}

		bestOrder := best.Value.(*Order)
		newTrade := bestOrder.Match(newOrder)

		if newTrade == nil {
			break
		}

		trades = append(trades, newTrade)

		if bestOrder.Filled() {
			makerBooks.Remove(bestOrder.Key())
		}

		if newOrder.Filled() {
			return trades
		}
	}

	if newOrder.Filled() {
		log.Fatalf("[oceanbook.orderbook] unexpected filled order %od", newOrder.ID)
	}

	takerBooks.Put(newOrder.Key(), newOrder)

	return trades
}

// CancelOrder removes order with specified id.
func (od *OrderBook) CancelOrder(o *Order) {
	od.Lock()
	defer od.Unlock()

	od.Bids.Remove(o.Key())
	od.Asks.Remove(o.Key())
}
