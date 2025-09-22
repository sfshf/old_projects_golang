package pswds

import (
	"errors"
	"log"
	"os"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"google.golang.org/grpc"
)

func DialDefault() (*grpc.ClientConn, error) {
	return DialConnectorGrpc(grpc.WithInsecure(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
}

func DialConnectorGrpc(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		err := errors.New("must set env variable for 'CONSUL_HTTP_ADDR'")
		log.Println(err)
		return nil, err
	}
	return grpc.Dial("consul://"+consulAddr+"/pswds", opts...)
}
