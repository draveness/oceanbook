package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:9121", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := oceanbookpb.NewOceanbookClient(conn)
	_, err = client.NewOrderBook(ctx, &oceanbookpb.NewOrderBookRequest{
		Pair: "BTC-USDT",
	})

	if err != nil {
		panic(err)

	}

	ordersCount := 10000

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(ordersCount)

	for i := 1; i <= ordersCount; i++ {
		var side oceanbookpb.Order_Side
		switch rand.Intn(2) {
		case 0:
			side = oceanbookpb.Order_ASK
		case 1:
			side = oceanbookpb.Order_BID
		}

		request := oceanbookpb.InsertOrderRequest{
			Id:       uint64(i),
			Pair:     "BTC-USDT",
			Side:     side,
			Price:    "1.0",
			Quantity: "2.0",
		}

		stream, err := client.InsertOrder(ctx, &request)
		if err != nil {
			log.Errorf("insert order %d error, err: %s", request.Id, err.Error())
		}

		go func() {
			for {
				trade, err := stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Fatalf("can not receive %v", err)
					return
				}

				log.Infof("receive trades %v", trade)
			}
		}()

		wg.Done()
	}
	time.Sleep(5 * time.Second)
	wg.Wait()
	fmt.Println(time.Since(start))
}
