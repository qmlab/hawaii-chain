package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	InitConfig("config.json")
	assert.True(t, len(Usrcfg.Address) > 0)
	assert.Equal(t, 100.0, InitialAccounts[0].Val)
}
