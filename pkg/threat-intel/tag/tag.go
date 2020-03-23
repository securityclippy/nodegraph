package tag

import "github.com/securityclippy/nodegraph/pkg/node"

type Tag struct {
	node.Node
	Tag string `json:"tag,omitempty"`
}
