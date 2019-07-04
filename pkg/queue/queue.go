package queue

import "github.com/draveness/oceanbook/pkg/order"

// OrderQueue is the FIFO queue to store orders.
type OrderQueue struct {
	size   int64
	values []*order.Order
}

// NewOrderQueue returns an order queue.
func NewOrderQueue(size int64) OrderQueue {
	return OrderQueue{
		size:   size,
		values: make([]*order.Order, 0, size),
	}
}

// Push appends an order to the end of order queue.
func (oq *OrderQueue) Push(o *order.Order) {
	oq.values = append(oq.values, o)
}

// First returns the first order in order queue.
func (oq *OrderQueue) First() *order.Order {
	if oq.Size() <= 0 {
		return nil
	}

	return oq.values[0]
}

// Pop removes and returns the first order in the order queue.
func (oq *OrderQueue) Pop() *order.Order {
	if oq.Size() <= 0 {
		return nil
	}

	o := oq.values[0]

	oq.values = oq.values[1:]
	return o
}

// Clear removes all orders.
func (oq *OrderQueue) Clear() {
	oq.values = make([]*order.Order, 0, oq.size)
}

// Values returns all orders.
func (oq *OrderQueue) Values() []*order.Order {
	return oq.values
}

// Size returns the size of orders in the queue.
func (oq *OrderQueue) Size() int64 {
	return int64(len(oq.values))
}
