// Package vfsutil provides utility functions for the VFS layer.
package vfsutil

import "github.com/rclone/rclone/vfs/vfscommon"

// NoVFSQuoteNames is a global flag that controls whether EscapeName
// applies quoting to filenames. When set to true, EscapeName returns
// the name unchanged (no quoting). This is used for compatibility
// with FileBrowser and similar tools that don't expect quoted names.
//
// This flag is set by the --no-vfs-quote-names command line option.
var NoVFSQuoteNames = false

// EscapeName wraps a filename with quotes for shell compatibility.
//
// When NoVFSQuoteNames is true (set via --no-vfs-quote-names flag),
// returns the original name unchanged.
//
// When NoVFSQuoteNames is false (default), returns the name with
// single quotes for shell compatibility.
func EscapeName(name string) string {
	if NoVFSQuoteNames || vfscommon.Opt.NoVFSQuoteNames {
		return name
	}
	// Default behavior: wrap with single quotes for shell compatibility
	return "'" + name + "'"
}
