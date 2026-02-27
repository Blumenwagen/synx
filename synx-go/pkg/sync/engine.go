package sync

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Blumenwagen/synx/pkg/config"
	"github.com/Blumenwagen/synx/pkg/ui"
)

type Engine struct {
	ConfigDir  string
	DotfileDir string
	Config     *config.ConfigManager
	DryRun     bool
}

func NewEngine(cfg *config.ConfigManager) (*Engine, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &Engine{
		ConfigDir:  filepath.Join(home, ".config"),
		DotfileDir: filepath.Join(home, "dotfiles"),
		Config:     cfg,
	}, nil
}

type SyncResult struct {
	Synced  int
	Skipped int
}

func (e *Engine) SyncToRepo() (*SyncResult, error) {
	if err := os.MkdirAll(e.DotfileDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dotfiles dir: %w", err)
	}

	res := &SyncResult{}

	for _, target := range e.Config.GetAllTargets() {
		srcPath := filepath.Join(e.ConfigDir, target)
		destPath := filepath.Join(e.DotfileDir, target)

		if e.Config.IsProfileTarget(target) {
			destPath = filepath.Join(e.DotfileDir, "profiles", e.Config.ActiveProfile, target)
		}

		resolvedSrc, err := filepath.EvalSymlinks(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				ui.Warn(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(not found)")))
				res.Skipped++
				continue
			}
			ui.Error(fmt.Sprintf("failed to resolve %s: %v", target, err))
			res.Skipped++
			continue
		}
		srcPath = resolvedSrc

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			ui.Warn(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(not found)")))
			res.Skipped++
			continue
		} else if err != nil {
			ui.Error(fmt.Sprintf("failed to stat %s: %v", target, err))
			res.Skipped++
			continue
		}

		if info.IsDir() {
			if e.DryRun {
				ui.Success(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(would sync)")))
				res.Synced++
			} else {
				err = e.copyDirWithExcludes(srcPath, destPath, target)
				if err != nil {
					ui.Error(fmt.Sprintf("%s %s", target, ui.StyleDim.Render(fmt.Sprintf("(copy failed: %v)", err))))
					res.Skipped++
				} else {
					res.Synced++
				}
			}
		} else {
			if e.Config.IsExcluded(target) {
				ui.Warn(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(excluded)")))
				res.Skipped++
				continue
			}
			if e.DryRun {
				ui.Success(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(would sync)")))
				res.Synced++
			} else {
				err = e.copyFile(srcPath, destPath)
				if err != nil {
					ui.Error(fmt.Sprintf("%s %s", target, ui.StyleDim.Render(fmt.Sprintf("(copy failed: %v)", err))))
					res.Skipped++
				} else {
					ui.Success(target)
					res.Synced++
				}
			}
		}
	}

	if !e.DryRun {
		synxDataDir := filepath.Join(e.DotfileDir, ".synx")
		os.MkdirAll(synxDataDir, 0755)

		bsSrc := filepath.Join(e.Config.ConfigDir, "bootstrap.conf")
		bsDst := filepath.Join(synxDataDir, "bootstrap.conf")
		if _, err := os.Stat(bsSrc); err == nil {
			e.copyFile(bsSrc, bsDst)
		}

		synxSrc := e.Config.SynxConfig
		synxDst := filepath.Join(synxDataDir, "synx.conf")
		if _, err := os.Stat(synxSrc); err == nil {
			e.copyFile(synxSrc, synxDst)
		}

		excSrc := e.Config.ExcludeCfg
		excDst := filepath.Join(synxDataDir, "exclude.conf")
		if _, err := os.Stat(excSrc); err == nil {
			e.copyFile(excSrc, excDst)
		}

		for _, name := range []string{"packages.native", "packages.foreign", "services.system", "services.user"} {
			src := filepath.Join(e.Config.ConfigDir, name)
			if _, err := os.Stat(src); err == nil {
				e.copyFile(src, filepath.Join(synxDataDir, name))
			}
		}
	}

	return res, nil
}

type RestoreResult struct {
	Restored int
	Failed   int
}

func (e *Engine) RestoreFromRepo() (*RestoreResult, error) {
	res := &RestoreResult{}

	for _, target := range e.Config.GetAllTargets() {
		srcPath := filepath.Join(e.DotfileDir, target)
		if e.Config.IsProfileTarget(target) {
			srcPath = filepath.Join(e.DotfileDir, "profiles", e.Config.ActiveProfile, target)
		}
		destPath := filepath.Join(e.ConfigDir, target)

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			ui.Warn(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(not in dotfiles repo)")))
			continue
		} else if err != nil {
			ui.Error(fmt.Sprintf("failed to stat %s in repo: %v", target, err))
			continue
		}

		destInfo, destErr := os.Lstat(destPath)

		if destErr == nil {
			if destInfo.Mode()&os.ModeSymlink != 0 {
				realPath, err := filepath.EvalSymlinks(destPath)
				if err != nil {
					ui.Error(fmt.Sprintf("failed to eval symlink %s: %v", target, err))
					res.Failed++
					continue
				}
				if info.IsDir() {
					os.MkdirAll(realPath, 0755)
				} else {
					os.MkdirAll(filepath.Dir(realPath), 0755)
				}
				err = e.restoreDirInPlace(srcPath, realPath, target)
				if err != nil {
					ui.Error(fmt.Sprintf("%s %s", target, ui.StyleDim.Render(fmt.Sprintf("(restore failed: %v)", err))))
					res.Failed++
				} else {
					ui.Success(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(symlink preserved)")))
					res.Restored++
				}
			} else if destInfo.IsDir() && info.IsDir() {
				os.MkdirAll(destPath, 0755)
				err = e.restoreDirInPlace(srcPath, destPath, target)
				if err != nil {
					ui.Error(fmt.Sprintf("%s %s", target, ui.StyleDim.Render(fmt.Sprintf("(restore failed: %v)", err))))
					res.Failed++
				} else {
					ui.Success(target)
					res.Restored++
				}
			} else {
				if e.Config.IsExcluded(target) {
					ui.Warn(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(excluded)")))
					continue
				}

				err = e.copyFile(srcPath, destPath)
				if err != nil {
					ui.Error(fmt.Sprintf("%s %s", target, ui.StyleDim.Render(fmt.Sprintf("(restore failed: %v)", err))))
					res.Failed++
				} else {
					ui.Success(target)
					res.Restored++
				}
			}
		} else {
			if info.IsDir() {
				err = e.copyDirSimple(srcPath, destPath) // We don't need to filter on restore if it doesn't exist
				if err != nil {
					ui.Error(fmt.Sprintf("%s %s", target, ui.StyleDim.Render(fmt.Sprintf("(restore failed: %v)", err))))
					res.Failed++
				} else {
					ui.Success(target)
					res.Restored++
				}
			} else {
				if e.Config.IsExcluded(target) {
					ui.Warn(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(excluded)")))
					continue
				}
				os.MkdirAll(filepath.Dir(destPath), 0755)
				err = e.copyFile(srcPath, destPath)
				if err != nil {
					ui.Error(fmt.Sprintf("%s %s", target, ui.StyleDim.Render(fmt.Sprintf("(restore failed: %v)", err))))
					res.Failed++
				} else {
					ui.Success(target)
					res.Restored++
				}
			}
		}
	}

	return res, nil
}

// Helpers

func (e *Engine) copyDirWithExcludes(src, dst, targetName string) error {
	excludedCount := 0

	dst = resolveDestSymlink(dst)

	resolvedConfigDir, err := filepath.EvalSymlinks(e.ConfigDir)
	if err != nil {
		resolvedConfigDir = e.ConfigDir
	}

	tmpDst := dst + ".tmp"
	os.RemoveAll(tmpDst)

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(resolvedConfigDir, path)
		if err != nil {
			return err
		}

		if e.Config.IsExcluded(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			excludedCount++
			return nil
		}

		dstPath := filepath.Join(tmpDst, strings.TrimPrefix(path, src))

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if d.Type()&fs.ModeSymlink != 0 {
			linkDst, err := os.Readlink(path)
			if err == nil {
				os.Remove(dstPath)
				os.Symlink(linkDst, dstPath)
			}
			return nil
		}

		return e.copyFile(path, dstPath)
	})

	if err != nil {
		os.RemoveAll(tmpDst)
		return err
	}

	// Atomic swap
	backupDst := dst + ".backup"
	os.RemoveAll(backupDst)

	if _, statErr := os.Stat(dst); statErr == nil {
		os.Rename(dst, backupDst)
	}

	if err := os.Rename(tmpDst, dst); err != nil {
		os.Rename(backupDst, dst) // attempt restore
		os.RemoveAll(tmpDst)
		return fmt.Errorf("failed atomic directory rename: %w", err)
	}

	os.RemoveAll(backupDst)

	if excludedCount > 0 {
		ui.Success(fmt.Sprintf("%s %s", targetName, ui.StyleDim.Render(fmt.Sprintf("(%d excluded)", excludedCount))))
	} else {
		ui.Success(targetName)
	}

	return nil
}

func (e *Engine) restoreDirInPlace(src, dst, targetName string) error {
	excludedCount := 0

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(e.DotfileDir, path)
		if err != nil {
			return err
		}

		if e.Config.IsExcluded(relPath) {
			excludedCount++
			return nil
		}

		dstPath := filepath.Join(dst, strings.TrimPrefix(path, src))

		if d.Type()&fs.ModeSymlink != 0 {
			linkDst, err := os.Readlink(path)
			if err == nil {
				os.MkdirAll(filepath.Dir(dstPath), 0755)
				os.Remove(dstPath)
				os.Symlink(linkDst, dstPath)
			}
			return nil
		}

		dstPath = resolveDestSymlink(dstPath)

		os.MkdirAll(filepath.Dir(dstPath), 0755)
		return e.copyFile(path, dstPath)
	})

	if err != nil {
		return err
	}

	if excludedCount > 0 {
	}

	return nil
}

func (e *Engine) copyDirSimple(src, dst string) error {
	dst = resolveDestSymlink(dst)

	tmpDst := dst + ".tmp"
	os.RemoveAll(tmpDst)

	cmd := exec.Command("cp", "-r", src, tmpDst) // Use OS cp for quick recursive copy, preserve symlinks
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDst)
		return err
	}

	backupDst := dst + ".backup"
	os.RemoveAll(backupDst)

	if _, statErr := os.Stat(dst); statErr == nil {
		os.Rename(dst, backupDst)
	}

	if err := os.Rename(tmpDst, dst); err != nil {
		os.Rename(backupDst, dst)
		os.RemoveAll(tmpDst)
		return fmt.Errorf("failed atomic directory rename: %w", err)
	}

	os.RemoveAll(backupDst)
	return nil
}

func (e *Engine) copyFile(src, dst string) error {
	dst = resolveDestSymlink(dst)

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmpDst := dst + ".tmp"
	out, err := os.Create(tmpDst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		out.Close()
		os.Remove(tmpDst)
		return err
	}
	out.Close()

	info, err := os.Stat(src)
	if err == nil {
		os.Chmod(tmpDst, info.Mode())
	}

	if err := os.Rename(tmpDst, dst); err != nil {
		os.Remove(tmpDst)
		return fmt.Errorf("failed atomic file rename: %w", err)
	}

	return nil
}

// resolveDestSymlink follows a symlink at dst so we write through it
// rather than replacing it with a regular file.
func resolveDestSymlink(dst string) string {
	if linfo, err := os.Lstat(dst); err == nil && linfo.Mode()&os.ModeSymlink != 0 {
		if resolved, err := filepath.EvalSymlinks(dst); err == nil {
			return resolved
		}
	}
	return dst
}
