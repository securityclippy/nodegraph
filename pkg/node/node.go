package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	log "github.com/sirupsen/logrus"
)

type Node struct {
	//type of node
	Type  string   `json:"type"`
	//Name node
	Name  string   `json:"name"`
	//Node UID in graph
	UID   string   `json:"uid"`
	//
	Description string `json:"description"`
	//Byte representation of node
	Raw   []byte   `json:"raw"`
	Links []string `json:"links"`
	//Unique resource ID
	URID string `json:"urid"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func New(nodeType, name string) *Node {
	node := Node{
		Type:    nodeType,
		Name:    name,
		Created: time.Now(),
		Updated: time.Now(),
	}
	return &node
}

func (n *Node) Link(relationship string, n2 *Node, db *dgo.Dgraph) error {
	txn := db.NewTxn()
	defer txn.Discard(context.Background())

	if n2 == nil {
		log.Errorf("cannot create link to nil node")
	}

	if n2.UID == "" {
		return fmt.Errorf("n2 does not have uid.  Cannot create link")
	}

	edge := map[string]interface{}{
		"uid": n.UID,
		relationship: map[string]string{"uid": n2.UID},
	}

	out, err := json.Marshal(edge)
	if err != nil {
		return err
	}
	_, err = txn.Mutate(context.Background(), &api.Mutation{
		SetJson: out,
	})
	if err != nil {
		return err
	}
	return nil
}


// LinkAndUpsert links n to n2 via the relationship and upserts n2 if it does not exist
func (n *Node) UpsertAndLink(relationship string, n2 *Node, db *dgo.Dgraph) (*Node, *Node, error) {
	existingN2, err := n2.Existing(db)
	if err != nil {
		return n, existingN2, err
	}
	if existingN2 == nil {
		n2, err = n2.Upsert(db)
		if err != nil {
			return n, nil, err
		}
	}

	err = n.Link(relationship, n2, db)
	if err != nil {
		return n, n2, err
	}

	return n, n2, nil
}


func (n *Node) LinkMultiple(relationship string, nodes []*Node, db *dgo.Dgraph) []*error {
	txn := db.NewTxn()
	defer txn.Discard(context.Background())

	errs := []*error{}
	for _, n2 := range nodes {
		func() {
			if n2 == nil {
				err := errors.New("cannot create link to nil node")
				errs = append(errs, &err)
				return
			}

			if n2.UID == "" {
				err := errors.New("n2 does not have uid.  Cannot create link")
				errs = append(errs, &err)
				return
			}

			in := []byte(fmt.Sprintf(`{"uid": %q, %q: {"uid": %q}}`, n.UID, relationship, n2.UID))
			log.Info(string(in))
			raw := map[string]interface{}{}
			err := json.Unmarshal(in, &raw)
			if err != nil {
				errs = append(errs, &err)
				return
			}
			out, err := json.Marshal(raw)
			if err != nil {
				errs = append(errs, &err)
				return
			}
			_, err = txn.Mutate(context.Background(), &api.Mutation{
				SetJson: out,
			})
			if err != nil {
				errs = append(errs, &err)
				return
			}
		}()

	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (n *Node) Existing(db *dgo.Dgraph) (*Node, error) {
	ctx := context.Background()
	txn := db.NewTxn()
	defer txn.Discard(ctx)
	q := fmt.Sprintf(`{
			node(func: allofterms(type, %q)) @filter(eq(name, %q)) {
			name 
			uid
			}
			}`, n.Type, n.Name)
	var decode struct {
		nodes []*Node
	}
	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.GetJson(), &decode)
	if err != nil {
		log.Error(err)
	}
	if len(decode.nodes) > 0 {

		return decode.nodes[0], nil
	}
	return nil, nil
}

func (n *Node) Upsert(db *dgo.Dgraph) (*Node, error) {

	existing, err := n.Existing(db)
	if err != nil {
		return nil, err
	}
	// doesn't exist, just create it
	txn := db.NewTxn()
	defer txn.Discard(context.Background())
	if existing == nil {
		out, err := json.Marshal(n)
		if err != nil {
			return nil, err
		}
		resp, err := txn.Mutate(context.Background(), &api.Mutation{SetJson: out})
		n.UID = resp.Uids["blank-0"]
		return n, nil
	}
	return existing, nil
}


