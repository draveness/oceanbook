package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/draveness/oceanbook/api/protobuf-spec/oceanbookpb"
	_ "github.com/draveness/oceanbook/pkg/log"
	"github.com/draveness/oceanbook/pkg/service/oceanbook"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	svc := oceanbook.NewService()

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpcprometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpcprometheus.UnaryServerInterceptor),
	)

	oceanbookpb.RegisterOceanbookServer(grpcServer, svc)

	grpcprometheus.Register(grpcServer)

	var sigCh = make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		log.Infof("[oceanbook] received signal: %+v", sig)
		log.Infof("[oceanbook] gracefully shutdown oceanbook server")
		grpcServer.GracefulStop()
		log.Infof("[oceanbook] shutdown oceanbook server")
		os.Exit(0)
	}()

	log.Infof("[oceanbook] start oceanbook at port 9121...")
	listen, err := net.Listen("tcp", ":9121")
	if err != nil {
		log.Fatalf("[oceanbook] failed to listen: %v", err)
	}

	log.Infof("[oceanbook] ready to serve...")
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("[oceanbook] serve error, err: %s", err.Error())
	}
}
