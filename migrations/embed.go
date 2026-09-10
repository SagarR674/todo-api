// Package migrations embeds the versioned SQL migration files so the
// application and its tests can apply them without any external files or the
// golang-migrate CLI.
package migrations

import "embed"

// FS holds every *.sql migration in this directory.
//
//go:embed *.sql
var FS embed.FS
