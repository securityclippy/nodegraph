package client

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

func NewDgraphClient(host string) (*dgo.Dgraph, error) {
	if host == "" {
		host = "127.0.0.1:9080"
	}
	d, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := dgo.NewDgraphClient(api.NewDgraphClient(d))
	return c, nil
}



