package orderbook

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/emirpasic/gods/utils"
	log "github.com/sirupsen/logrus"
)

// OrderSide is the orders' side.
type OrderSide string

const (
	// OrderSideAsk represents the ask order side.
	OrderSideAsk OrderSide = "ask"

	// OrderSideBid represents the bid order side.
	OrderSideBid OrderSide = "bid"
)

// Order .
type Order struct {
	ID                uint64          `json:"id"`
	Side              OrderSide       `json:"side"`
	Price             decimal.Decimal `json:"price"`
	Quantity          decimal.Decimal `json:"quantity"`
	FilledQuantity    decimal.Decimal `json:"filled_quantity"`
	StopPrice         decimal.Decimal `json:"stop_price"`
	CreatedAt         time.Time       `json:"created_at"`
	ImmediateOrCancel bool            `json:"immediate_or_cancel"`
}

// OrderKey is used to sort orders in red black tree.
type OrderKey struct {
	ID        uint64          `json:"id"`
	Side      OrderSide       `json:"side"`
	Price     decimal.Decimal `json:"price"`
	CreatedAt time.Time       `json:"created_at"`
}

// Key returns a OrderKey.
func (o *Order) Key() *OrderKey {
	return &OrderKey{
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
func (o *Order) Match(taker *Order) *Trade {
	maker := o
	if maker.Side == taker.Side {
		log.Fatalf("[oceanbook.orderbook] match order with same side %d, %d", maker.ID, taker.ID)
		return nil
	}

	var bidOrder *Order
	var askOrder *Order

	switch maker.Side {
	case OrderSideBid:
		bidOrder = maker
		askOrder = taker

	case OrderSideAsk:
		bidOrder = taker
		askOrder = maker
	}

	switch {
	case taker.IsLimit():
		if bidOrder.Price.GreaterThanOrEqual(askOrder.Price) {
			filledQuantity := decimal.Min(bidOrder.PendingQuantity(), askOrder.PendingQuantity())
			bidOrder.Fill(filledQuantity)
			askOrder.Fill(filledQuantity)

			return &Trade{
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

		return &Trade{
			Price:    maker.Price,
			Quantity: filledQuantity,
			TakerID:  taker.ID,
			MakerID:  maker.ID,
		}
	}

	return nil
}

// OrderComparator is used for comparing OrderKey.
func OrderComparator(a, b interface{}) (result int) {
	this := a.(*OrderKey)
	that := b.(*OrderKey)

	if this.Side != that.Side {
		log.Fatalf("[oceanbook.orderbook] compare order with different sides")
	}

	if this.ID == that.ID {
		return
	}

	// based on ask
	switch {
	case this.Side == OrderSideAsk && this.Price.LessThan(that.Price):
		result = 1

	case this.Side == OrderSideAsk && this.Price.GreaterThan(that.Price):
		result = -1

	case this.Side == OrderSideBid && this.Price.LessThan(that.Price):
		result = -1

	case this.Side == OrderSideBid && this.Price.GreaterThan(that.Price):
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
