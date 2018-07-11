package merkle

// This is a package for string compression/decompression
import (
	"bytes"
	"compress/zlib"
)

func ZipBytes(bs []byte) ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write(bs)
	w.Close()
	if err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}

func ZipString(input string) (string, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write([]byte(input))
	w.Close()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func UnzipBytes(bs []byte) ([]byte, error) {
	b := bytes.NewReader(bs)
	r, err := zlib.NewReader(b)
	if err != nil {
		return []byte{}, err
	}
	defer r.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.Bytes(), nil
}

func UnzipString(input string) (string, error) {
	b := bytes.NewReader([]byte(input))
	r, err := zlib.NewReader(b)
	if err != nil {
		return "", err
	}
	defer r.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), nil
}
