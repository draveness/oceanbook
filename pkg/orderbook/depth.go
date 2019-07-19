package orderbook

import (
	"github.com/draveness/oceanbook/pkg/order"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// PriceLevel .
type PriceLevel struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
	Side     order.Side
	Count    uint64
}

// PriceLevelKey .
type PriceLevelKey struct {
	Price decimal.Decimal
	Side  order.Side
}

// Key returns a key for PriceLevel.
func (pl *PriceLevel) Key() *PriceLevelKey {
	return &PriceLevelKey{
		Price: pl.Price,
		Side:  pl.Side,
	}
}

// Depth .
type Depth struct {
	Pair  string
	Scale int64
	Bids  *rbt.Tree
	Asks  *rbt.Tree
}

// NewDepth returns a depth with specific scale.
func NewDepth(pair string, scale int64) *Depth {
	return &Depth{
		Pair:  pair,
		Scale: scale,
		Bids:  rbt.NewWith(PriceLevelComparator),
		Asks:  rbt.NewWith(PriceLevelComparator),
	}
}

// UpdatePriceLevel updates depth with price level.
func (d *Depth) UpdatePriceLevel(pl *PriceLevel) {
	var priceLevels *rbt.Tree

	switch pl.Side {
	case order.SideAsk:
		priceLevels = d.Asks

	case order.SideBid:
		priceLevels = d.Bids

	default:
		log.Fatalf("[depth] invalid price level side %s", pl.Side)
	}

	foundPriceLevel, found := priceLevels.Get(pl.Key())
	if !found {
		priceLevels.Put(pl.Key(), pl)
		return
	}

	existedPriceLevel := foundPriceLevel.(*PriceLevel)
	existedPriceLevel.Quantity = existedPriceLevel.Quantity.Add(pl.Quantity)
	existedPriceLevel.Count += pl.Count

	if existedPriceLevel.Count == 0 || existedPriceLevel.Quantity.Equal(decimal.Zero) {
		priceLevels.Remove(existedPriceLevel.Key())
	}
}

// PriceLevelComparator .
func PriceLevelComparator(a, b interface{}) int {
	this := a.(*PriceLevelKey)
	that := b.(*PriceLevelKey)

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
