package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"pet/middleware/class"
	"pet/services"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 51051, "The service port")
)

func main() {
	flag.Parse()
	err := godotenv.Load(".env", ".env.local")
	if err != nil {
		log.Println("Warning loading .env file")
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	cache, _ := services.NewDefaultCache()
	server := &service{db: NewDatabaseClass(), cache: cache}
	class.RegisterServiceServer(grpcServer, server)
	log.Printf("server listening at %v", listen.Addr())
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
