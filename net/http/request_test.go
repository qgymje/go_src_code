package http_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFmtQ(t *testing.T) {
	s := "hello"
	fmt.Printf("%q", s)
}

func TestRemoveZone(t *testing.T) {
	s := "[fe80::1%en0]:80080"
	v := removeZone(s)
	assert.Equal(t, v, "[fe80::1]:80080", "equal")
}
