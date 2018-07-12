package utils

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"math/rand"
	"proto"
	"time"

	"github.com/gogo/protobuf/proto"
)

func GetHash(n *pb.Node) (string, error) {
	if n.Count == 0 && len(n.Val) == 0 {
		return "", nil
	}
	return Hash(n)
}

func Hash(msg proto.Message) (string, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return "", err
	}

	return HashBytes(data), nil
}

func HashBytes(data []byte) string {
	h := sha256.New()
	h.Write(data)
	hs := b64.StdEncoding.EncodeToString(h.Sum(nil))
	return hs
}

func ToNibbles(s string) []byte {
	var rst []byte
	bs := []byte(s)
	for i := 0; i < len(s); i++ {
		rst = append(rst, bs[i]/16)
		rst = append(rst, bs[i]%16)
	}
	return rst
}

func ToInts(bs []byte) []int {
	var rst []int
	for _, b := range bs {
		rst = append(rst, int(b))
	}
	return rst
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
