package grpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	pb "github.com/blokhinnv/shorty/proto"
)

// withServerUnaryInterceptor returns unary intercept options.
func withServerUnaryInterceptor(srv *ShortyServer) grpc.ServerOption {
	return grpc.UnaryInterceptor(srv.userTokenInceptor)
}

// withServerStreamInterceptor returns unary intercept options.
func withServerStreamInterceptor(srv *ShortyServer) grpc.ServerOption {
	return grpc.StreamInterceptor(srv.userTokenStreamInterceptor)
}

// RunGRPCServer creates the store and starts the server.
func RunGRPCServer(ctx context.Context, cfg *config.ServerConfig) {
	// defining the port for the server
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	// creating a gRPC server without a registered service
	s, err := database.NewDBStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}
	srvCloseCh := make(chan struct{}, 1)
	srvImpl := NewShortyServer(
		s,
		cfg.BaseURL,
		[]byte(cfg.SecretKey),
		cfg.TrustedSubnet,
		1*time.Second,
		srvCloseCh,
	)
	srv := grpc.NewServer(withServerUnaryInterceptor(srvImpl), withServerStreamInterceptor(srvImpl))
	// registering the service

	pb.RegisterShortyServer(srv, srvImpl)
	log.Printf("Starting gRPC server with config %+v\n", cfg)
	// getting a gRPC request
	go func() {
		if err := srv.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	// signal to finish deleting goroutines
	srvCloseCh <- struct{}{}
	log.Println("Shutting down server gracefully...")
	// ensure goroutines are finished
	<-srvCloseCh
	log.Println("Bye!")
}
