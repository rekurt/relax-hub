//go:build tools
// +build tools

package tools

// Tool dependencies - imported to keep in go.mod until actual usage in code.
// Remove this file once internal/admin/ imports these packages directly.
import (
	_ "github.com/GoAdminGroup/go-admin/adapter/chi"
	_ "github.com/GoAdminGroup/go-admin/engine"
	_ "github.com/GoAdminGroup/themes/adminlte"
)
