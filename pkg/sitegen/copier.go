package sitegen

import (
	"errors"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"io"
	"os"
	"path/filepath"
)

func copyStaticFiles() error {
	staticDir := filepath.Join(config.Cfg.TemplatesDir, "static")
	dstDir := filepath.Join(config.Cfg.Output.Webroot, "static")

	if info, err := os.Stat(staticDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Nothing to copy
			return nil
		}
		if !info.IsDir() {
			return errors.New("static directory is not a directory")
		}
	}

	if err := copyDir(staticDir, dstDir); err != nil {
		return err
	}
	return nil

}

// CopyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Sync()
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
