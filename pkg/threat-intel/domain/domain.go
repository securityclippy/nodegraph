package domain

import (
	"github.com/securityclippy/nodegraph/pkg/node"
)


var Schema = `

	domain: string @index(term, fulltext) .
	registered_with: [uid] @reverse .
	hosted_by: [uid] @reverse .
	type domain {
		domain
		name
		address: ipaddress
	}

`

type DomainNode struct {
	*node.Node
	Domain string `json:"domain"`
	Address string `json:"address,omitempty"`
}



func New(domain string) *DomainNode {
	return &DomainNode{
		Domain:domain,
		Node: node.New("domain", domain, ""),
	}
}
