package pages

import (
	"encoding/binary"
	"encoding/gob"
	"io"
	"os"
	"structure-lite/internal/errs"
)

type Page[T any] struct {
	file    *os.File
	decoder *gob.Decoder
	encoder *gob.Encoder
}

func NewPage[T any](file *os.File) *Page[T] {
	return &Page[T]{
		file:    file,
		decoder: gob.NewDecoder(file),
		encoder: gob.NewEncoder(file),
	}
}

func Open[T any](name string) (*Page[T], error) {
	const op errs.Op = "page.Open"

	file, err := os.OpenFile(name, os.O_RDWR, 0666)
	if err != nil {
		return nil, errs.With(op, "unable to open file", err)
	}

	return NewPage[T](file), nil
}

func Create[T any](name string) (*Page[T], error) {
	const op errs.Op = "page.Create"

	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, errs.With(op, "unable to create file", err)
	}

	page := NewPage[T](file)

	if err = page.writeCount(0); err != nil {
		return nil, errs.Wrap(op, err)
	}

	return page, nil
}

func (p *Page[T]) ReadAllItems(dst *[]T) error {
	const op errs.Op = "page.ReadAllItems"

	if err := p.skipHeader(); err != nil {
		return errs.Wrap(op, err)
	}

	for {
		item := new(T)
		err := p.decoder.Decode(item)

		if err == io.EOF {
			break
		}

		if err != nil {
			return errs.Wrap(op, err)
		}

		*dst = append(*dst, *item)
	}

	return nil
}

func (p *Page[T]) ReadCount() (int, error) {
	const op errs.Op = "page.ReadCount"

	var count int32

	if _, err := p.file.Seek(0, io.SeekStart); err != nil {
		return 0, errs.Wrap(op, err)
	}

	if err := binary.Read(p.file, binary.LittleEndian, &count); err != nil {
		return 0, errs.Wrap(op, err)
	}

	return int(count), nil
}

func (p *Page[T]) InsertItem(item any) error {
	const op errs.Op = "page.InsertItem"

	if err := p.incrementCount(); err != nil {
		return errs.Wrap(op, err)
	}

	if err := p.addItemToEnd(item); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (p *Page[T]) Close() error {
	const op errs.Op = "page.Close"

	if err := p.file.Close(); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (p *Page[T]) writeCount(count int) error {
	const op errs.Op = "page.writeCount"

	if _, err := p.file.Seek(0, io.SeekStart); err != nil {
		return errs.Wrap(op, err)
	}

	if err := binary.Write(p.file, binary.LittleEndian, int32(count)); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (p *Page[T]) Trunc() error {
	const op errs.Op = "page.Trunc"

	if err := p.file.Truncate(0); err != nil {
		return errs.Wrap(op, err)
	}

	if err := p.writeCount(0); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (p *Page[T]) Name() string {
	return p.file.Name()
}

func (p *Page[T]) incrementCount() error {
	const op errs.Op = "page.incrementCount"

	var count int
	var err error

	if count, err = p.ReadCount(); err != nil {
		return errs.Wrap(op, err)
	}

	count++

	if err = p.writeCount(count); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (p *Page[T]) skipHeader() error {
	const op errs.Op = "page.SkipHeader"

	const int32size = 4

	if _, err := p.file.Seek(int32size, io.SeekStart); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (p *Page[T]) addItemToEnd(item any) error {
	const op errs.Op = "page.addItemToEnd"

	if _, err := p.file.Seek(0, io.SeekEnd); err != nil {
		return errs.Wrap(op, err)
	}

	if err := p.encoder.Encode(item); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}
