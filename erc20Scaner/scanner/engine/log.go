package engine

import "log/slog"

// log is the package-level logger used by Process and Client methods.
// Must be set via SetLogger before calling any Process methods.
var log *slog.Logger

// SetLogger sets the package-level structured logger.
func SetLogger(l *slog.Logger) {
	log = l
}
