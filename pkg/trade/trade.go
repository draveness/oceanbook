package trade

import (
	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	"github.com/shopspring/decimal"
)

// Trade .
type Trade struct {
	ID       uint64
	Symbol   string
	Price    decimal.Decimal
	Quantity decimal.Decimal
	TakerID  uint64
	MakerID  uint64
}

// Serialize returns protobuf encoded trade.
func (t *Trade) Serialize() *oceanbookpb.Trade {
	return &oceanbookpb.Trade{
		Id:       t.ID,
		Symbol:   t.Symbol,
		Price:    t.Price.String(),
		Quantity: t.Quantity.String(),
		TakerId:  t.TakerID,
		MakerId:  t.MakerID,
	}
}
