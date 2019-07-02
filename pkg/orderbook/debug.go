package orderbook

import (
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
)

func (od *OrderBook) String() string {
	var askStrs []string
	var bidStrs []string

	maxLength := 0
	for _, ask := range od.Asks.Values() {
		askOrder := ask.(*Order)
		askStr := fmt.Sprintf("%s %10s * %-10s %10s %s\n",
			aurora.Red("|"),
			askOrder.Price.StringFixed(6),
			askOrder.PendingQuantity().StringFixed(6),
			askOrder.Price.Mul(askOrder.Quantity).StringFixed(6),
			aurora.Red("|"))
		askStrs = append(askStrs, askStr)
		if len(askStr) > maxLength {
			maxLength = len(askStr)
		}
	}

	for _, bid := range od.Bids.Values() {
		bidOrder := bid.(*Order)
		bidStr := fmt.Sprintf("%s %10s * %-10s %10s %s\n",
			aurora.Green("|"),
			bidOrder.Price.StringFixed(6),
			bidOrder.PendingQuantity().StringFixed(6),
			bidOrder.Price.Mul(bidOrder.Quantity).StringFixed(6),
			aurora.Green("|"))
		bidStrs = append(bidStrs, bidStr)
		if len(bidStr) > maxLength {
			maxLength = len(bidStr)
		}
	}

	str := fmt.Sprintf("OrderBook: %s\n", od.Market)

	var border string
	if maxLength > 19 {
		border = fmt.Sprintf("%s\n", strings.Repeat("-", maxLength-19))
	}

	str += fmt.Sprint(aurora.Red(border))
	for _, askStr := range askStrs {
		str += askStr
	}
	str += fmt.Sprint(aurora.Red(border))

	str += fmt.Sprintln()

	str += fmt.Sprint(aurora.Green(border))
	for _, bidStr := range bidStrs {
		str += bidStr
	}
	str += fmt.Sprint(aurora.Green(border))

	return str
}
