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
	"runtime/debug"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 51051, "The service port")
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Recovered from panic",
				slog.String("stack", string(debug.Stack())),
				slog.String("err", fmt.Sprintf("%v", err)))
		}
	}()
	ctx := context.Background()
	go services.Handle(ctx, func(context.Context) {
		slog.Info("Service exit")
		os.Exit(0)
	})
	services.DefineLogging()
	flag.Parse()
	err := godotenv.Load(".env", ".env.local")
	if err != nil {
		slog.Warn("Warning loading .env file", slog.String("err", err.Error()))
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		slog.Error("Failed to listen", slog.String("err", err.Error()))
		return
	}
	grpcServer := grpc.NewServer()
	cache, _ := services.NewDefaultCache(ctx)
	db := NewDatabaseClass()
	_ = db.CreateClass("main", "Main")
	server := &service{db: db, cache: cache}
	class.RegisterServiceServer(grpcServer, server)
	slog.Info("Starting server", slog.String("addr", listen.Addr().String()))
	if err = grpcServer.Serve(listen); err != nil {
		slog.Error("Failed to serve ", slog.String("err", err.Error()))
	}
}
