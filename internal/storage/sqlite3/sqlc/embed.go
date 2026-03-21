package sqlc

import "embed"

//go:embed schema/*.sql
var SqlFiles embed.FS
