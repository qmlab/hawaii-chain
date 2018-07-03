package utils

import (
	"crypto/sha1"
	b64 "encoding/base64"
	"merkle/proto"

	"github.com/gogo/protobuf/proto"
)

func GetHash(n *pb.Node) (string, error) {
	// Count is reference count for the number of children the current node has
	// If Count == 0, it means there is no child. If this is not a leaf node (Val==""), it is safe to remove the node
	// If Count == 1, it means it is a pass-through node without branching (assuming Val==""), it is safe to compact the node
	if n.Count <= 1 && len(n.Val) == 0 {
		return "", nil
	}
	data, err := proto.Marshal(n)
	if err != nil {
		return "", err
	}

	h := sha1.New()
	h.Write(data)
	hs := b64.StdEncoding.EncodeToString(h.Sum(nil))
	return hs, err
}
