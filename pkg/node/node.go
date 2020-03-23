package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/securityclippy/nodegraph/pkg/edge"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	log "github.com/sirupsen/logrus"
)

type Node struct {
	//type of node
	DType  string   `json:"dgraph.type,omitempty"`
	//Name node
	Name  string   `json:"name,omitempty"`
	//Node UID in graph
	UID   string   `json:"uid,omitempty"`
	//
	Description string `json:"description,omitempty"`
	//Byte representation of node
	Raw   []byte   `json:"raw,omitempty"`
	//Links []string `json:"links"`
	//Unique resource ID
	//URID string `json:"urid"`

	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

func New(nodeType, name, uid string) *Node {
	if uid == "" {
		return 	&Node{
			DType:    nodeType,
			Name:    name,
			Created: time.Now(),
			Updated: time.Now(),
		}
	}

	return &Node{
		UID: uid,
		DType:    nodeType,
		Name:    name,
		Created: time.Now(),
		Updated: time.Now(),
	}
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

	//edge := map[string]interface{}{
		//"uid": n.UID,
		//relationship: map[string]string{"uid": n2.UID},
	//}

	e := edge.New(n.UID, relationship, n2.UID)

	out, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = txn.Mutate(context.Background(), &api.Mutation{
		SetJson: out,
		CommitNow:true,
	})
	if err != nil {
		return err
	}
	return nil
}


// LinkAndUpsert links n to n2 via the relationship and upserts n2 if it does not exist
func (n *Node) UpsertAndLink(relationship string, n2 *Node, db *dgo.Dgraph) (node *Node, existingNode *Node, err error) {
	existingNode, err = n2.Existing(db)
	if err != nil {
		return n, existingNode, err
	}
	if existingNode == nil {
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


//func (n *Node) LinkMulti(relationship string, nodes []*Node, db *dgo.Dgraph) error {
	//for _, n := range nodes {
		//
	//}
//}

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
				CommitNow:true,
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
			}`, n.DType, n.Name)
	var decode struct {
		Node []struct{
			Name string `json:"name"`
			UID string `json:"uid"`
		}
	}

	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.GetJson(), &decode)

	if err != nil {
		log.Error(err)
	}
	if len(decode.Node) > 0 {

		n := &Node{
			Name: decode.Node[0].Name,
			UID: decode.Node[0].UID,
		}
		return n, nil
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
		resp, err := txn.Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow:true})
		for _, v := range resp.GetUids() {
			n.UID = v
		}
		return n, nil
	}
	return existing, nil
}

func (n *Node) JSONString() string {
	js, _ := json.MarshalIndent(n, "", "  ")
	return string(js)
}

func (n *Node) Set(db *dgo.Dgraph) (*Node, error) {
	out, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	txn := db.NewTxn()
	defer txn.Discard(ctx)

	resp, err := txn.Mutate(context.Background(), &api.Mutation{SetJson:out, CommitNow:true})

	if err != nil {
		return nil, err
	}

	for _, v := range resp.GetUids() {
		n.UID = v
	}

	return n, nil
}
