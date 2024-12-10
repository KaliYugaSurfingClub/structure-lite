package pages

import (
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
