package orderbook

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/emirpasic/gods/utils"
	log "github.com/sirupsen/logrus"
)

type OrderSide string

const (
	OrderSideAsk OrderSide = "ask"
	OrderSideBid OrderSide = "bid"
)

type Order struct {
	ID             uint64          `json:"id"`
	Side           OrderSide       `json:"side"`
	Price          decimal.Decimal `json:"price"`
	Quantity       decimal.Decimal `json:"quantity"`
	FilledQuantity decimal.Decimal `json:"filled_quantity"`
	StopPrice      decimal.Decimal `json:"stop_price"`
	CreatedAt      time.Time       `json:"created_at"`
}

type OrderKey struct {
	ID        uint64          `json:"id"`
	Side      OrderSide       `json:"side"`
	Price     decimal.Decimal `json:"price"`
	CreatedAt time.Time       `json:"created_at"`
}

func (o *Order) Key() *OrderKey {
	return &OrderKey{
		ID:        o.ID,
		Side:      o.Side,
		Price:     o.Price,
		CreatedAt: o.CreatedAt,
	}
}

func (o *Order) Filled() bool {
	return o.Quantity.Equal(o.FilledQuantity)
}

func (o *Order) PendingQuantity() decimal.Decimal {
	return o.Quantity.Sub(o.FilledQuantity)
}

func (o *Order) Fill(quantity decimal.Decimal) {
	o.FilledQuantity = o.FilledQuantity.Add(quantity)
}

func (o *Order) IsLimit() bool {
	return o.Price.IsPositive()
}

func (o *Order) IsMarket() bool {
	return o.Price.IsZero()
}

func (maker *Order) Match(taker *Order) *Trade {
	if maker.Side == taker.Side {
		log.Fatalf("[oceanbook.orderbook] match order with same side %d, %d", maker.ID, taker.ID)
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

func OrderComparator(a, b interface{}) (result int) {
	this := a.(*OrderKey)
	that := b.(*OrderKey)

	if this.Side != that.Side {
		log.Fatalf("[oceanbook.orderbook] compare order with different sides")
	}

	if this.ID == that.ID {
		result = 0
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
