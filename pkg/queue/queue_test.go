package queue

import (
	"testing"

	"github.com/draveness/oceanbook/pkg/order"
	"github.com/stretchr/testify/suite"
)

type OrderQueueTestSuite struct {
	suite.Suite
}

func (s *OrderQueueTestSuite) TestOrderQueue() {
	orderQueue := NewOrderQueue(5)

	s.Nil(orderQueue.First())
	s.Nil(orderQueue.Pop())

	for i := uint64(0); i < 10; i++ {
		orderQueue.Push(&order.Order{
			ID: i,
		})

		s.Equal(int64(i+1), orderQueue.Size())
	}

	for i := uint64(0); i < 10; i++ {
		s.Equal(&order.Order{ID: i}, orderQueue.First())
		s.Equal(&order.Order{ID: i}, orderQueue.Pop())
		s.Equal(int64(9-i), orderQueue.Size())
	}
}

func TestOrderQueue(t *testing.T) {
	suite.Run(t, new(OrderQueueTestSuite))
}
