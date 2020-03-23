package registrar

import "github.com/securityclippy/nodegraph/pkg/node"


var Schema = `

	registered: [uid] @reverse .
	registrar: string @index(term, fulltext) .

	type registrar {
		name
		registrar
	}

`

type RegistrarNode struct {
	*node.Node
	Registrar string `json:"registrar,omitemtpy"`
	Country string `json:"country,omitempty"`
}


func New(registrar string) *RegistrarNode {
	return &RegistrarNode{
		Registrar:registrar,
		Node: node.New("registrar", registrar, ""),
	}
}
