package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"pet/middleware/class"
	"pet/services"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 51051, "The service port")
)

func main() {
	services.DefineLogging()
	flag.Parse()
	err := godotenv.Load(".env", ".env.local")
	if err != nil {
		slog.Warn("Warning loading .env file", slog.String("error", err.Error()))
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		slog.Error("Failed to listen", slog.String("error", err.Error()))
		os.Exit(1)
	}
	grpcServer := grpc.NewServer()
	cache, _ := services.NewDefaultCache(context.Background())
	server := &service{db: NewDatabaseClass(), cache: cache}
	class.RegisterServiceServer(grpcServer, server)
	slog.Debug("Server listening at ", slog.String("address", listen.Addr().String()))
	if err = grpcServer.Serve(listen); err != nil {
		slog.Error("Failed to serve ", slog.String("error", err.Error()))
	}
}
