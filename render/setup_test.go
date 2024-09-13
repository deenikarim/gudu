package render

import (
	"github.com/CloudyKit/jet/v6"
	"os"
	"testing"
)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader(""),
	jet.InDevelopmentMode())

var testRender = Render{
	RendererEngine:    "",
	TemplatesRootPath: "",
	Port:              "",
	JetViews:          views,
	DevelopmentMode:   false,
}

// TestMain
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
