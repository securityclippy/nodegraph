package nodeops

import (
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/securityclippy/nodegraph/pkg/edge"
	"github.com/securityclippy/nodegraph/pkg/node"
	"context"
	"encoding/json"
)

func BulkAddNodes(nodes []*node.Node, db *dgo.Dgraph) (map[string]string, error) {
	txn := db.NewTxn()
	defer txn.Discard(context.Background())

	/*for _, n := range nodes {

		out, err := json.Marshal(n)
		if err != nil {
			return nil, err
		}
		resp, err := txn.Mutate(context.Background(), &api.Mutation{SetJson: out})
	}*/
	out, err := json.Marshal(nodes)
	if err != nil {
		return nil, err
	}
	resp, err := txn.Mutate(context.Background(), &api.Mutation{SetJson: out})
	if err != nil {
		return nil, err
	}
	return resp.Uids, nil
}

func BulkLink(rootNode *node.Node, relationship string, childUIDS map[string]string, db *dgo.Dgraph) error {
	txn := db.NewTxn()
	defer txn.Discard(context.Background())

	edges := []*edge.Edge{}
	for _, u := range childUIDS {
		e := edge.New(rootNode.UID, relationship, u)
		edges = append(edges, &e)
	}
	out, err := json.Marshal(edges)
	if err != nil {
		return err
	}
	_, err = txn.Mutate(context.Background(), &api.Mutation{SetJson: out})
	if err != nil {
		return err
	}
	return nil
}
