package static

import "embed"

var (
	//go:embed all:static
	FileSystem embed.FS
)

const (
	RootPath  = "static"
	IndexPath = RootPath + "/index.html"
)
