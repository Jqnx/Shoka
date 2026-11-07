package handlers

import (
	"os"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

// The gallery handlers render through libvips, so the package's test binary
// needs the library initialized once.
func TestMain(m *testing.M) {
	_ = vips.Startup(nil)
	code := m.Run()

	vips.Shutdown()

	os.Exit(code)
}
