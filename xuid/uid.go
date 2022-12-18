package xuid

import (
	"github.com/bwmarrin/snowflake"
	"github.com/rs/xid"
)

//xuid作为主键

// snowflake (8bytes)

type Node struct {
	*snowflake.Node
}

// nodeID : 1 --> 1024

func NewNode(nodeID int64) (*Node, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}
	return &Node{node}, nil
}

func (n *Node) GenerateID() int64 {
	return n.Generate().Int64()
}

// xid (12bytes)

func GenerateID() string {
	return xid.New().String()
}
