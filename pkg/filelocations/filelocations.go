// Documents important file paths used by the system
package filelocations

import (
	"path/filepath"

	"github.com/joonas-fi/joonas-sys/pkg/common"
)

var (
	Sysroot = WithRoot("/sysroot")
)

// use non-default sysroot location.
// this is roughly analogous to `$prefix` in filesystem hierarchies.
func WithRoot(root string) Root {
	return Root{root}
}

// represents base path like "/sysroot" under which all subdirs are resolved from.
// this is roughly analogous to `$prefix` in filesystem hierarchies.
type Root struct {
	root string // most times "/sysroot"
}

// "/sysroot/apps/OS-checkout/<sysid>"
func (b Root) Checkout(sysID string) string {
	return filepath.Join(b.CheckoutsDir(), sysID)
}

// "/sysroot/apps/OS-checkout"
func (b Root) CheckoutsDir() string {
	return b.App(common.AppOSCheckout)
}

// "/sysroot/apps/<appName>"
func (b Root) App(appName string) string {
	return filepath.Join(b.root, "apps", appName)
}

// "/sysroot/apps/OS-diff/<sysID>"
func (b Root) Diff(sysID string) string {
	return filepath.Join(b.App(common.AppOSDiff), sysID)
}

// "/sysroot"
// please use only as-is (don't use this to derive other values)
func (b Root) Root() string {
	return b.root
}
