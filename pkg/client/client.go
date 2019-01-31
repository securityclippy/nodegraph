package client

import (
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
	"log"
)

func NewDgraphClient(host string) *dgo.Dgraph {
	if host == "" {
		host = "localhost:9080"
	}
	d, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c := dgo.NewDgraphClient(api.NewDgraphClient(d))
	return c
}

