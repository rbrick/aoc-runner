package ui

import (
	"embed"
	"io/fs"
)

//go:generate npm ci
//go:generate npm run build

//go:embed all:dist
var dist embed.FS

// the built web interface, run `go generate ./ui` to build it
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")

	if err != nil {
		panic(err)
	}

	return sub
}

func Built() bool {
	_, err := fs.Stat(dist, "dist/index.html")
	return err == nil
}
