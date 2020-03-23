package person

import (
	"fmt"
	"github.com/securityclippy/nodegraph/pkg/node"
)

var Schema = `
	
	registered: [uid] @reverse .
	first_name: string @index(term, fulltext) .
	last_name: string @index(term, fulltext) .	

	type person {
		name
		first_name
		last_name
	}

`

type PersonNode struct {
	*node.Node
	FirstName string `json:"first_name,omitempty"`
	LastName string `json:"last_name,omitempty"`
}

func New(first, last string) *PersonNode {
	return &PersonNode{
		FirstName:first,
		LastName:last,
		Node: node.New("person", fmt.Sprintf("%s %s", first, last), ""),
	}
}