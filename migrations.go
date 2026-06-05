package url_shortener

import "embed"

//go:embed migrations
var EmbeddedMigrations embed.FS
