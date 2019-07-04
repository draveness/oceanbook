package order

import (
	"time"

	"github.com/draveness/oceanbook/pkg/trade"
	"github.com/emirpasic/gods/utils"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// Side is the orders' side.
type Side string

const (
	// SideAsk represents the ask order side.
	SideAsk Side = "ask"

	// SideBid represents the bid order side.
	SideBid Side = "bid"
)

// Order .
type Order struct {
	ID                uint64          `json:"id"`
	Side              Side            `json:"side"`
	Price             decimal.Decimal `json:"price"`
	StopPrice         decimal.Decimal `json:"stop_price"`
	Quantity          decimal.Decimal `json:"quantity"`
	FilledQuantity    decimal.Decimal `json:"filled_quantity"`
	CreatedAt         time.Time       `json:"created_at"`
	ImmediateOrCancel bool            `json:"immediate_or_cancel"`
}

// Key is used to sort orders in red black tree.
type Key struct {
	ID        uint64          `json:"id"`
	Side      Side            `json:"side"`
	Price     decimal.Decimal `json:"price"`
	StopPrice decimal.Decimal `json:"stop_price"`
	CreatedAt time.Time       `json:"created_at"`
}

// Key returns a Key.
func (o *Order) Key() *Key {
	return &Key{
		ID:        o.ID,
		Side:      o.Side,
		Price:     o.Price,
		CreatedAt: o.CreatedAt,
	}
}

// Filled returns true when its filled quantity equals to quantity.
func (o *Order) Filled() bool {
	return o.Quantity.Equal(o.FilledQuantity)
}

// PendingQuantity is the remaing quantity.
func (o *Order) PendingQuantity() decimal.Decimal {
	return o.Quantity.Sub(o.FilledQuantity)
}

// Fill updates order filled quantity with passing arguments.
func (o *Order) Fill(quantity decimal.Decimal) {
	o.FilledQuantity = o.FilledQuantity.Add(quantity)
}

// IsLimit returns true when the order is limit order.
func (o *Order) IsLimit() bool {
	return o.Price.IsPositive()
}

// IsMarket returns true when the order is market order.
func (o *Order) IsMarket() bool {
	return o.Price.IsZero()
}

// Match matches maker with a taker and returns trade if there is a match.
func (o *Order) Match(taker *Order) *trade.Trade {
	maker := o
	if maker.Side == taker.Side {
		log.Fatalf("[oceanbook.orderbook] match order with same side %s, %d, %d", maker.Side, maker.ID, taker.ID)
		return nil
	}

	var bidOrder *Order
	var askOrder *Order

	switch maker.Side {
	case SideBid:
		bidOrder = maker
		askOrder = taker

	case SideAsk:
		bidOrder = taker
		askOrder = maker
	}

	switch {
	case taker.IsLimit():
		if bidOrder.Price.GreaterThanOrEqual(askOrder.Price) {
			filledQuantity := decimal.Min(bidOrder.PendingQuantity(), askOrder.PendingQuantity())
			bidOrder.Fill(filledQuantity)
			askOrder.Fill(filledQuantity)

			return &trade.Trade{
				Price:    maker.Price,
				Quantity: filledQuantity,
				TakerID:  taker.ID,
				MakerID:  maker.ID,
			}
		}

		return nil

	case taker.IsMarket():
		filledQuantity := decimal.Min(bidOrder.PendingQuantity(), askOrder.PendingQuantity())
		bidOrder.Fill(filledQuantity)
		askOrder.Fill(filledQuantity)

		return &trade.Trade{
			Price:    maker.Price,
			Quantity: filledQuantity,
			TakerID:  taker.ID,
			MakerID:  maker.ID,
		}
	}

	return nil
}

// Comparator is used for comparing Key.
func Comparator(a, b interface{}) (result int) {
	this := a.(*Key)
	that := b.(*Key)

	if this.Side != that.Side {
		log.Fatalf("[oceanbook.orderbook] compare order with different sides")
	}

	if this.ID == that.ID {
		return
	}

	// based on ask
	switch {
	case this.Side == SideAsk && this.Price.LessThan(that.Price):
		result = 1

	case this.Side == SideAsk && this.Price.GreaterThan(that.Price):
		result = -1

	case this.Side == SideBid && this.Price.LessThan(that.Price):
		result = -1

	case this.Side == SideBid && this.Price.GreaterThan(that.Price):
		result = 1

	default:
		switch {
		case this.CreatedAt.Before(that.CreatedAt):
			result = 1

		case this.CreatedAt.After(that.CreatedAt):
			result = -1

		default:
			result = utils.UInt64Comparator(this.ID, that.ID) * -1
		}
	}

	return
}

// StopComparator is used for comparing Key.
func StopComparator(a, b interface{}) (result int) {
	this := a.(*Key)
	that := b.(*Key)

	if this.Side != that.Side {
		log.Fatalf("[oceanbook.orderbook] compare order with different sides")
	}

	if this.ID == that.ID {
		return
	}

	// based on ask
	switch {
	case this.Side == SideAsk && this.StopPrice.LessThan(that.StopPrice):
		result = 1

	case this.Side == SideAsk && this.StopPrice.GreaterThan(that.StopPrice):
		result = -1

	case this.Side == SideBid && this.StopPrice.LessThan(that.StopPrice):
		result = -1

	case this.Side == SideBid && this.StopPrice.GreaterThan(that.StopPrice):
		result = 1

	default:
		switch {
		case this.CreatedAt.Before(that.CreatedAt):
			result = 1

		case this.CreatedAt.After(that.CreatedAt):
			result = -1

		default:
			result = utils.UInt64Comparator(this.ID, that.ID) * -1
		}
	}

	return
}
