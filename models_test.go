package goscope2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMessageHash(t *testing.T) {
	expect := "098f6bcd46"
	result := generateMessageHash("test")
	assert.Equal(t, expect, result)
}
