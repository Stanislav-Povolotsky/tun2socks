package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHijackConfig(t *testing.T) {
	t.Cleanup(DisableHijack)

	assert.Equal(t, "", HijackTarget())
	assert.False(t, IsHijackQuery(53))

	EnableHijack("8.8.8.8:53")
	assert.Equal(t, "8.8.8.8:53", HijackTarget())
	assert.True(t, IsHijackQuery(53))
	assert.False(t, IsHijackQuery(80), "only port 53 should be hijacked")

	DisableHijack()
	assert.Equal(t, "", HijackTarget())
	assert.False(t, IsHijackQuery(53))
}
