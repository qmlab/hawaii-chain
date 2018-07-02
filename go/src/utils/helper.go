package utils

import (
	"crypto/sha1"
	b64 "encoding/base64"
	"merkle/proto"

	"github.com/gogo/protobuf/proto"
)

func GetHash(n *pb.Node) (string, error) {
	if n.Count == 0 && len(n.Val) == 0 {
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
