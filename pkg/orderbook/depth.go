package orderbook

import (
	"github.com/draveness/oceanbook/pkg/order"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/shopspring/decimal"
)

// PriceLevel .
type PriceLevel struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
	Side     order.Side
	Count    uint64
}

// Depth .
type Depth struct {
	Pair  string
	Scale int64
	Bids  *treeset.Set
	Asks  *treeset.Set
}

// NewDepth returns a depth with specific scale.
func NewDepth(pair string, scale int64) *Depth {
	return &Depth{
		Pair:  pair,
		Scale: scale,
		Bids:  treeset.NewWith(PriceLevelComparator),
		Asks:  treeset.NewWith(PriceLevelComparator),
	}
}

// PriceLevelComparator .
func PriceLevelComparator(a, b interface{}) int {
	this := a.(*PriceLevel)
	that := b.(*PriceLevel)

	switch {
	case this.Side == order.SideAsk && this.Price.LessThan(that.Price):
		return 1

	case this.Side == order.SideAsk && this.Price.GreaterThan(that.Price):
		return -1

	case this.Side == order.SideBid && this.Price.LessThan(that.Price):
		return -1

	case this.Side == order.SideBid && this.Price.GreaterThan(that.Price):
		return 1

	default:
	}

	return 0
}
