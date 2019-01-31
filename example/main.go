package main

import (
	"bufio"
	"fmt"

	//"fmt"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/securityclippy/nodegraph/pkg/client"
	"github.com/securityclippy/nodegraph/pkg/node"
	"github.com/securityclippy/nodegraph/pkg/nodeops"
	"log"
	"os"
	"strings"
	"context"
)


var schema = `
			user: string @index(term) .
			role: string @index(term) .
			group: string @index(term) .
			permission: string @index(term) .
			iamrole: string @index(term) .
			type: string @index(term) .
			name: string @index(term) .
			policy: uid @reverse .
			allows: uid @reverse .
`

func ReadPerms() []string {
	perms := []string{}
	file, err := os.Open("aws_permissions.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() != "" {
			perms = append(perms, scanner.Text())
		}
	}
	return perms
}

func permRoot(perm string) string {
	return strings.Split(perm, ":")[0]
}

func dropDB(db *dgo.Dgraph) {
	db.Alter(context.Background(), &api.Operation{DropAll: true})
}


func SetSchema(schema string, db *dgo.Dgraph) error {

	err := db.Alter(context.Background(), &api.Operation{
		Schema: schema,
	})
	if err != nil {
		return err
	}
	return nil

}

func main() {
	db := client.NewDgraphClient("")

	dropDB(db)

	if err := SetSchema(schema, db); err != nil {
		log.Println(err)
	}

	perms := ReadPerms()

	permRootMap := map[string][]string{}


	for _, p := range perms {
		root := permRoot(p)
		rootStar := fmt.Sprintf("%s:*", root)
		_, ok := permRootMap[rootStar]
		if !ok {
			permRootMap[rootStar] = []string{}

		} else {
			permRootMap[rootStar] = append(permRootMap[rootStar], p)
		}
	}
	for k, v := range permRootMap {
		n1, _ := node.New("action", k).Upsert(db)
		nodes := []*node.Node{}
		for _, val := range v {
			n := node.New("action", val)
			nodes = append(nodes, n)
		}
		uids, err := nodeops.BulkAddNodes(nodes, db)
		if err != nil {
			log.Println(err)
		}
		err = nodeops.BulkLink(n1, "allows", uids, db)
		if err != nil {
			log.Println(err)
		}
		log.Printf("uploaded %d nodes", len(uids))
	}
}



