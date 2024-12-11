package pages

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"structure-lite/internal/errs"
)

func tempFileFrom(file *os.File) (*os.File, error) {
	const op errs.Op = "page.incrementItemsCount"

	dir := filepath.Dir(file.Name())

	tmp, err := os.CreateTemp(dir, "tmpfile-*")

	if err != nil {
		return nil, errs.W(op, err)
	}

	if _, err = io.Copy(tmp, file); err != nil {
		return nil, errs.W(op, err)
	}

	return tmp, nil
}

func DeleteIfErr[T any](err error, page *Page[T]) func() {
	return func() {
		if err == nil {
			return
		}

		deleteError := page.Delete()

		if deleteError != nil {
			err = errors.Join(err, deleteError)
		}
	}
}
