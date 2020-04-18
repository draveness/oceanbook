package orderbook

import (
	"sync"

	// log level and settings
	_ "github.com/draveness/oceanbook/pkg/log"
	"github.com/draveness/oceanbook/pkg/order"
	"github.com/draveness/oceanbook/pkg/queue"
	"github.com/draveness/oceanbook/pkg/trade"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// OrderBook is the order book.
type OrderBook struct {
	sync.RWMutex
	Symbol string
	Price  decimal.Decimal

	Bids     *rbt.Tree
	Asks     *rbt.Tree
	StopBids *rbt.Tree
	StopAsks *rbt.Tree

	pendingOrdersQueue *queue.OrderQueue
	cancelOrdersQueue  map[uint64]*order.Order

	depth *Depth
}

const (
	// pendingOrdersCap is the buffer size for pending orders.
	pendingOrdersCap int64 = 1024
)

// NewOrderBook returns a pointer to an orderbook.
func NewOrderBook(symbol string) *OrderBook {
	orderQueue := queue.NewOrderQueue(pendingOrdersCap)
	return &OrderBook{
		Symbol:             symbol,
		Bids:               rbt.NewWith(order.Comparator),
		Asks:               rbt.NewWith(order.Comparator),
		StopBids:           rbt.NewWith(order.StopComparator),
		StopAsks:           rbt.NewWith(order.StopComparator),
		pendingOrdersQueue: &orderQueue,
		cancelOrdersQueue:  make(map[uint64]*order.Order, 1024),
		depth:              NewDepth(symbol, 16),
	}
}

// InsertOrder inserts new order into orderbook.
func (od *OrderBook) InsertOrder(newOrder *order.Order) []*trade.Trade {
	od.Lock()
	defer od.Unlock()

	log.Debugf("[oceanbook.orderbook] insert order with id %d - %s * %s, side %s", newOrder.ID, newOrder.Price, newOrder.Quantity, newOrder.Side)

	if !newOrder.StopPrice.Equal(decimal.Zero) {
		od.insertStopOrder(newOrder)

		return []*trade.Trade{}
	}

	trades := od.insertOrder(newOrder)

	pendingOrders := od.pendingOrdersQueue.Values()
	for i := range pendingOrders {
		pendingOrder := pendingOrders[i]

		log.Debugf("[oceanbook.orderbook] insert stop order with id %d - %s * %s, side %s", pendingOrder.ID, pendingOrder.Price, pendingOrder.Quantity, pendingOrder.Side)

		newTrades := od.insertOrder(pendingOrder)
		trades = append(trades, newTrades...)
	}
	od.pendingOrdersQueue.Clear()

	return trades
}

func (od *OrderBook) insertOrder(newOrder *order.Order) []*trade.Trade {
	trades := []*trade.Trade{}

	var takerBooks, makerBooks *rbt.Tree
	switch newOrder.Side {
	case order.SideAsk:
		takerBooks = od.Asks
		makerBooks = od.Bids

	case order.SideBid:
		takerBooks = od.Bids
		makerBooks = od.Asks

	default:
		log.Fatalf("[oceanbook.orderbook] invalid order side %s", newOrder.Side)
		return trades
	}

	_, found := takerBooks.Get(newOrder.Key())
	if found {
		return trades
	}

	for {
		if newOrder == nil {
			break
		}

		best := makerBooks.Right()
		if best == nil {
			break
		}

		bestOrder := best.Value.(*order.Order)
		newTrade := bestOrder.Match(newOrder)

		if newTrade == nil {
			break
		}

		trades = append(trades, newTrade)
		log.Debugf("[oceanbook.orderbook] new trade %d with price %s", newTrade.ID, newTrade.Price)

		od.depth.UpdatePriceLevel(&PriceLevel{
			Side:  bestOrder.Side,
			Price: newTrade.Quantity.Neg(),
		})

		if bestOrder.Filled() {
			makerBooks.Remove(bestOrder.Key())
			delete(od.cancelOrdersQueue, bestOrder.ID)
		}

		od.setMarketPrice(newTrade.Price)

		if newOrder.Filled() {
			return trades
		}
	}

	// if the order is immediate or cancel order, it is not supposed to insert
	// into the orderbooks.
	if newOrder.ImmediateOrCancel {
		return trades
	}

	od.depth.UpdatePriceLevel(&PriceLevel{
		Side:  newOrder.Side,
		Price: newOrder.PendingQuantity(),
	})
	takerBooks.Put(newOrder.Key(), newOrder)
	od.cancelOrdersQueue[newOrder.ID] = newOrder

	return trades
}

func (od *OrderBook) insertStopOrder(newOrder *order.Order) {
	var takerBooks *rbt.Tree
	switch newOrder.Side {
	case order.SideAsk:
		takerBooks = od.StopAsks

	case order.SideBid:
		takerBooks = od.StopBids

	default:
		log.Fatalf("[oceanbook.orderbook] invalid stop order side %s", newOrder.Side)
		return
	}

	_, found := takerBooks.Get(newOrder.Key())
	if found {
		return
	}

	takerBooks.Put(newOrder.Key(), newOrder)
}

func (od *OrderBook) setMarketPrice(newPrice decimal.Decimal) {
	previousPrice := od.Price
	od.Price = newPrice

	if previousPrice.Equal(decimal.Zero) {
		return
	}

	switch {
	case newPrice.LessThan(previousPrice):
		// price gone done, check stop asks
		for {
			best := od.StopBids.Right()
			if best == nil {
				break
			}

			bestOrder := best.Value.(*order.Order)
			if bestOrder.StopPrice.LessThan(newPrice) {
				break
			}

			log.Debugf("[oceanbook.orderbook] bid order %d with stop price %s enqueued", bestOrder.ID, bestOrder.Price)

			od.StopBids.Remove(best.Key)
			od.pendingOrdersQueue.Push(bestOrder)
		}

	case newPrice.GreaterThan(previousPrice):
		// price gone done, check stop asks
		for {
			best := od.StopAsks.Right()
			if best == nil {
				break
			}

			bestOrder := best.Value.(*order.Order)
			if bestOrder.StopPrice.GreaterThan(newPrice) {
				break
			}

			log.Debugf("[oceanbook.orderbook] ask order %d with stop price %s enqueued", bestOrder.ID, bestOrder.Price)

			od.StopAsks.Remove(best.Key)
			od.pendingOrdersQueue.Push(bestOrder)
		}

	default:
		// previous price equals to new price
		return
	}
}

// CancelOrder removes order with specified id.
func (od *OrderBook) CancelOrder(o *order.Order) {
	od.Lock()
	defer od.Unlock()

	targetOrder, ok := od.cancelOrdersQueue[o.ID]
	if !ok {
		return
	}

	switch targetOrder.Side {
	case order.SideAsk:
		od.Asks.Remove(targetOrder.Key())

	case order.SideBid:
		od.Bids.Remove(targetOrder.Key())

	default:
		od.Asks.Remove(targetOrder.Key())
		od.Bids.Remove(targetOrder.Key())
	}
}

// GetDepth returns the order book depth.
func (od *OrderBook) GetDepth() *Depth {
	od.RLock()
	defer od.RUnlock()
	return od.depth
}
