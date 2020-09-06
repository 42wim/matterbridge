package bmatrix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainUsername(t *testing.T) {
	uut := newMatrixUsername("MyUser")

	assert.Equal(t, "MyUser", uut.formatted)
	assert.Equal(t, "MyUser", uut.plain)
}

func TestHTMLUsername(t *testing.T) {
	uut := newMatrixUsername("<b>MyUser</b>")

	assert.Equal(t, "<b>MyUser</b>", uut.formatted)
	assert.Equal(t, "MyUser", uut.plain)
}

func TestFancyUsername(t *testing.T) {
	uut := newMatrixUsername("<MyUser>")

	assert.Equal(t, "&lt;MyUser&gt;", uut.formatted)
	assert.Equal(t, "<MyUser>", uut.plain)
}
