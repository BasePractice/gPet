package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"pet/middleware/class"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 51051, "The service port")
)

type service struct {
	class.UnimplementedServiceServer
	db DatabaseClass
}

func (s *service) Information(_ context.Context, request *class.InformationRequest) (*class.InformationReply, error) {
	log.Printf("Received: %+v\n", request)
	return &class.InformationReply{Version: 1}, nil
}

func main() {
	flag.Parse()
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	server := &service{db: NewDatabaseClass()}
	class.RegisterServiceServer(grpcServer, server)
	log.Printf("server listening at %v", listen.Addr())
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
