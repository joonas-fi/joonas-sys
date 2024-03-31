// Documents important file paths used by the system
package filelocations

import (
	"path/filepath"
)

var (
	Sysroot = WithRoot("/sysroot")
)

func WithRoot(root string) Root {
	return Root{root}
}

// represents base path like "/sysroot" under which all subdirs are resolved from
type Root struct {
	root string // most times "/sysroot"
}

func (b Root) CheckoutsDir() string {
	return b.App("OS-checkout")
}

func (b Root) App(appName string) string {
	return filepath.Join(b.root, "apps", appName)
}

// please use only as-is (don't use this to derive other values)
func (b Root) Root() string {
	return b.root
}
