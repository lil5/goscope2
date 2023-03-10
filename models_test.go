package goscope2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMessageHash(t *testing.T) {
	expect := "098f6bcd46"
	log := Goscope2Log{Message: "test"}
	log.GenerateHash()
	result := log.Hash
	assert.Equal(t, expect, result)
}
