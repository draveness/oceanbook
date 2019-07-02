package orderbook

import rbt "github.com/emirpasic/gods/trees/redblacktree"

type GroupedOrder struct {
	Pair     string
	Count    uint
	Price    float64
	Quantity float64
}

type GroupedOrderbook struct {
	Pair  string
	Scale int
	Bids  *rbt.Tree
	Asks  *rbt.Tree
}
