package actor

import "github.com/securityclippy/nodegraph/pkg/node"

type Actor struct {
	*node.Node
	CodeName string `json:"code_name,omitempty"`
}
