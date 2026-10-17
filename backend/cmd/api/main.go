package main

import (
	"backend/internal/memberships"
	"backend/internal/orgs"
	"backend/internal/roles"
	"backend/internal/teams"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	// shared by the HTTP and gRPC servers
	memSvc := memberships.NewService(pool)

	// HTTP: 8080
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           newMux(pool, memSvc),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Println("http listening on :8080")
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err,
			http.ErrServerClosed) {
			log.Printf("http error: %v", err)
			stop()
		}
	}()

	// gRPC: 8067
	lis, err := net.Listen("tcp", ":8067")
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}
	grpcSrv := grpc.NewServer()
	memberships.NewGRPCServer(memSvc).Register(grpcSrv)
	reflection.Register(grpcSrv)
	go func() {
		log.Println("grpc listening on :8067")
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("grpc error: %v", err)
			stop()
		}
	}()

	// shutdown
	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(),
		10*time.Second)
	defer cancel()

	grpcSrv.GracefulStop()
	return srv.Shutdown(shutdownCtx)
}

func newMux(pool *pgxpool.Pool, memSvc *memberships.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

	orgs.NewHandler(orgs.NewService(pool)).Routes(mux)
	teams.NewHandler(teams.NewService(pool)).Routes(mux)
	roles.NewHandler(roles.NewService(pool)).Routes(mux)
	memberships.NewHandler(memSvc).Routes(mux)

	return mux
}
