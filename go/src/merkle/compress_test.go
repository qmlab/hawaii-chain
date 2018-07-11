package merkle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZipBytes(t *testing.T) {
	orig := []byte("super string 123")
	comp, err := ZipBytes(orig)
	assert.Nil(t, err)
	uncomp, err := UnzipBytes(comp)
	assert.Nil(t, err)
	assert.Equal(t, orig, uncomp)
}

func TestZipString(t *testing.T) {
	orig := "foo string 123"
	comp, err := ZipString(orig)
	assert.Nil(t, err)
	uncomp, err := UnzipString(comp)
	assert.Nil(t, err)
	assert.Equal(t, orig, uncomp)
}
