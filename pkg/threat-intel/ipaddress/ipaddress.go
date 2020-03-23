package ipaddress

import (
	"github.com/securityclippy/nodegraph/pkg/node"
	//"github.com/securityclippy/nodegraph/pkg/threat-intel/domain"
)

var Schema = `
	
	hosts: [uid] @reverse .
	address: string @index(term, fulltext) .
	type ipaddress {
		address
		name
		hosts: domain
	}

`

type IPNode struct {
	*node.Node
	Address string `json:"address"`
	//Hosts []domain.DomainNode `json:"hosts,omitempty"`
}

func New(address string) *IPNode {
	return &IPNode{
		Address:address,
		Node: node.New("ipaddress", address, ""),
	}
}
