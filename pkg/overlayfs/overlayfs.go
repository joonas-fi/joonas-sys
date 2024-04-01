package overlayfs

import (
	"io/fs"
	"os"
)

// TODO: we have better implementation in lsw

type overlayFs struct {
	upper fs.FS
	lower fs.FS
}

// makes an "overlay FS", where files are accessed from "upper" (usually R/W) dir first, and if not exists,
// then from "lower" (usually read-only). for semantics see https://wiki.archlinux.org/index.php/Overlay_filesystem
func New(upper fs.FS, lower fs.FS) fs.FS {
	return &overlayFs{upper, lower}
}

func (o *overlayFs) Open(name string) (fs.File, error) {
	upperFile, err := o.upper.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return o.lower.Open(name)
		} else {
			return nil, err
		}
	}

	return upperFile, nil
}
