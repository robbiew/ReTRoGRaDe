package filesystem

import "regexp"

var filenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// SanitizeFilename replaces unsafe characters in the provided string with underscores. This keeps
// filenames compatible across filesystems while preserving the general shape of the original value.
func SanitizeFilename(name string) string {
	return filenameSanitizer.ReplaceAllString(name, "_")
}
