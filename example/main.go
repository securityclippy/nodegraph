package main

import (
	"bufio"
	"fmt"
	"github.com/securityclippy/nodegraph/pkg/nodeops"

	//"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"

	"github.com/securityclippy/nodegraph/pkg/client"
	"github.com/securityclippy/nodegraph/pkg/node"
	"context"
	log "github.com/sirupsen/logrus"
	//"github.com/securityclippy/nodegraph/pkg/nodeops"
	"os"
	"strings"
)


var schema = `
			user: string @index(term,fulltext) .
			role: string @index(term,fulltext) .
			group: string @index(term,fulltext) .
			permission: string @index(term,fulltext) .
			iamrole: string @index(term,fulltext) .
			type: string @index(term,fulltext) .
			name: string @index(term,fulltext) .
			policy: [uid] @reverse .
			allows: [uid] @reverse .
			attached: [uid] @reverse .
			grants: [uid] @reverse .
			member: [uid] @reverse .
			resource: [uid] @reverse . `

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

func AddRole(db *dgo.Dgraph) error {
	ngPolicy, err := node.New("policy", "s3-policy").Existing(db)
	if err != nil {
		return err
	}

	n, err := node.New("iamrole", "ec2-role").Upsert(db)
	if err != nil {
		return err
	}

	return n.Link("attached", ngPolicy, db)
}

func addAdminPolicy(db *dgo.Dgraph) (*node.Node, error) {

	pol, err := node.New("policy", "adminPolicy").Upsert(db)
	if err != nil {
		return nil, err
	}

	s3Star, err := node.New("action", "s3:*").Existing(db)
	if err != nil {
		return nil, err
	}

	iamStar, err := node.New("action", "iam:*").Upsert(db)
	if err != nil {
		return nil, err
	}

	err = pol.Link("attached", s3Star, db)
	if err != nil {
		return nil, err
	}
	err = pol.Link("attached", iamStar, db)
	if err != nil {
		return nil, err
	}

	return pol, nil

}

func addAdminUser(db *dgo.Dgraph) (*node.Node, error) {

	user, err := node.New("user", "adminUser").Upsert(db)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func addAdminGroup(db *dgo.Dgraph) (*node.Node, error) {
	adminGroup, err := node.New("group", "adminGroup").Upsert(db)
	if err != nil {
		return nil, err
	}

	return adminGroup, nil
}







func AddPolicy(db *dgo.Dgraph) error {
	s3Star, err := node.New("action", "s3:*").Existing(db)
	if err != nil {
		return err
	}

	n, err := node.New("policy", "s3-policy").Upsert(db)

	if err != nil {
		return err
	}

	err = n.Link("grants", s3Star, db)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	db, err := client.NewDgraphClient("")
	if err != nil {
		log.Fatal(err)
	}

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
		n1, err := node.New("action", k).Upsert(db)
		if err != nil {
			log.Error(err)
		}
		//fmt.Printf("N1: %+v", n1)
		nodes := []*node.Node{}
		for _, val := range v {
			nodes = append(nodes, node.New("action", val))
			//n2, err := node.New("action", val).Upsert(db)

			//fmt.Printf("n1(%s) -> %s -> n2(%s)\n", n1.UID, "allows",  n2.UID)
			//if err != nil {
				//log.Error(err)
			//}
			//err = n1.Link("allows", n2, db)
			//if err != nil {
				//log.Error(err)
			//}
		}
		uids, err := nodeops.BulkAddNodes(nodes, db)
		if err != nil {
			log.Println(err)
		}
		err = nodeops.BulkLink(n1, "allows", uids, db)
		if err != nil {
			log.Println(err)
//
		}
		log.Printf("uploaded %d nodes", len(uids))
	}

	/*
	err = AddPolicy(db)
	if err != nil {
		log.Fatal(err)
	}

	err = AddRole(db)
	if err != nil {
		log.Fatal(err)
	}

	adminPol, err := addAdminPolicy(db)
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = adminPol.UpsertAndLink("allows", node.New("resource", "s3-bucket"), db)

	adminUser, err := addAdminUser(db)
	if err != nil {
		log.Fatal(err)
	}

	adminGroup, err := addAdminGroup(db)
	if err != nil {
		log.Fatal(err)
	}

	err = adminGroup.Link("attached", adminPol, db)
	if err != nil {
		log.Fatalf("attached: %s", err.Error())
	}

	err = adminGroup.Link("member", adminUser, db)
	if err != nil {
		log.Fatalf("addMember: %s", err.Error())
	}

	_, _, err = adminGroup.UpsertAndLink("member", node.New("user", "adminUser2"), db)
	if err != nil {
		log.Fatalf("addMember: %s", err.Error())
	}
	 */

}
