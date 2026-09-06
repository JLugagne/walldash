package frontend_test

import (
	"io/fs"
	"testing"

	"github.com/JLugagne/ha-dash/frontend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrontendFS(t *testing.T) {
	assets := frontend.FS()
	require.NotNil(t, assets, "frontend.FS() must not be nil")

	// Verify index.html exists in embedded filesystem
	data, err := fs.ReadFile(assets, "index.html")
	require.NoError(t, err, "index.html must exist in embedded FS")
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "<div id=\"root\"></div>")
}
