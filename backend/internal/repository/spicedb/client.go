// Package spicedb connects to the shared SpiceDB instance.
package spicedb

import (
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(endpoint, presharedKey string) (*authzed.Client, error) {
	return authzed.NewClient(
		endpoint,
		grpcutil.WithInsecureBearerToken(presharedKey),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
