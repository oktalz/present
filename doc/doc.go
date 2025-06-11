package doc

import (
	_ "embed"
)

// Logo is the ASCII art logo for the present application.
//
//go:embed logo.txt
var Logo string
