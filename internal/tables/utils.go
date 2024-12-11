package tables

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"structure-lite/internal/errs"
	"structure-lite/internal/pages"
)

func ttype[T any]() string {
	return reflect.TypeOf((*T)(nil)).Elem().String()
}

func forPages[T any](path string, f func(page *pages.Page[T]) error) error {
	const op errs.Op = "table.forPages"

	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if entry.IsDir() {
			return nil
		}

		page, err := pages.Open[T](path)
		if err != nil {
			return fmt.Errorf("error opening page %s: %w", path, err)
		}

		if err = f(page); err != nil {
			return fmt.Errorf("error processing page %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return errs.W(op, err)
	}

	return nil
}
