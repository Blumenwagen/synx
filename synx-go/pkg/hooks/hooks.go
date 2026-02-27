package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Blumenwagen/synx/pkg/ui"
)

// Hook names
const (
	PreSync     = "pre-sync"
	PostSync    = "post-sync"
	PreRestore  = "pre-restore"
	PostRestore = "post-restore"
	PreProfile  = "pre-profile"
	PostProfile = "post-profile"
)

// RunHook executes a hook script if it exists and is executable.
// hooksDir is typically ~/.config/synx/hooks/
// Optional args are passed to the hook script.
// Returns nil if the hook doesn't exist (not an error).
func RunHook(hooksDir, hookName string, args ...string) error {
	scriptPath := filepath.Join(hooksDir, hookName)

	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		return nil // No hook, that's fine
	}
	if err != nil {
		return fmt.Errorf("hook stat error: %w", err)
	}

	if info.Mode()&0111 == 0 {
		ui.Warn(fmt.Sprintf("Hook '%s' exists but is not executable (chmod +x it)", hookName))
		return nil
	}

	ui.SubStep(fmt.Sprintf("Running %s hook...", hookName))

	cmd := exec.Command(scriptPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook '%s' failed: %w", hookName, err)
	}

	ui.Success(fmt.Sprintf("Hook %s completed", hookName))
	return nil
}
