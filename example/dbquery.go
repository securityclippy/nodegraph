package main

import (
	"github.com/securityclippy/nodegraph/pkg/client"
	"github.com/securityclippy/nodegraph/pkg/node"
	log "github.com/sirupsen/logrus"
)

func main() {
	db, err := client.NewDgraphClient("")

	existing, err := node.New("action", "s3:*").Existing(db)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Existing: %s", existing.JSONString())
}


