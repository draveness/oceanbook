package main

import (
	"net"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	_ "github.com/draveness/oceanbook/pkg/log"
	"github.com/draveness/oceanbook/pkg/service/oceanbook"
)

func main() {
	stopCh := make(chan struct{}, 1)
	svc := oceanbook.NewService(stopCh)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpcprometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpcprometheus.UnaryServerInterceptor),
	)

	oceanbookpb.RegisterOceanbookServer(grpcServer, svc)

	grpcprometheus.Register(grpcServer)

	listen, err := net.Listen("tcp", ":9121")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalln(err.Error())
	}
}
