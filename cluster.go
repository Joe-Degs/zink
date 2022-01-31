package zinc

type Status bool

const (
	ACTIVE   Status = true
	INACTIVE Status = false
)

type Node struct {
	*Peer
	Stat Status `json:"status"`
}

type Cluster struct {
	*Peer
	members []*Node
}
