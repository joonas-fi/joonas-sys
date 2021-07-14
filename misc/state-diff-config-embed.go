package misc

import (
	"embed"
)

//go:embed state-diff-config.hcl state-diff-ignore-temp.txt
var Files embed.FS
