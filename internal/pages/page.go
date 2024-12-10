package pages

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"structure-lite/internal/errs"
)

type Page[T any] struct {
	file    *os.File
	decoder *gob.Decoder
	encoder *gob.Encoder
}

func newPage[T any](file *os.File) *Page[T] {
	return &Page[T]{
		file:    file,
		decoder: gob.NewDecoder(file),
		encoder: gob.NewEncoder(file),
	}
}

func Open[T any](name string) (*Page[T], error) {
	const op errs.Op = "page.Open"

	file, err := os.OpenFile(name, os.O_RDWR, 0755)
	if err != nil {
		return nil, errs.W(op, err)
	}

	return newPage[T](file), nil
}

func Create[T any](name string) (*Page[T], error) {
	const op errs.Op = "page.Create"

	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, errs.W(op, err)
	}

	page := &Page[T]{file: file}

	if err = page.writeCount(0); err != nil {
		return nil, errs.W(op, err)
	}

	return newPage[T](file), nil
}

func (p *Page[T]) ReadAllItems(dst *[]T) error {
	const op errs.Op = "page.ReadAllItems"
	const int32size = 4

	if _, err := p.file.Seek(int32size, io.SeekStart); err != nil {
		return errs.W(op, err)
	}

	for {
		item := new(T)
		err := p.decoder.Decode(item)

		if err == io.EOF {
			break
		}

		if err != nil {
			return errs.W(op, err)
		}

		*dst = append(*dst, *item)
	}

	return nil
}

func (p *Page[T]) ReadCount() (int, error) {
	const op errs.Op = "page.ReadCount"

	var count int32

	if _, err := p.file.Seek(0, io.SeekStart); err != nil {
		return 0, errs.W(op, err)
	}

	if err := binary.Read(p.file, binary.LittleEndian, &count); err != nil {
		return 0, errs.W(op, err)
	}

	fmt.Println(op, count)

	return int(count), nil
}

func (p *Page[T]) writeCount(count int) error {
	const op errs.Op = "page.writeCount"

	if _, err := p.file.Seek(0, io.SeekStart); err != nil {
		return errs.W(op, err)
	}

	if err := binary.Write(p.file, binary.LittleEndian, int32(count)); err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (p *Page[T]) InsertItem(item any) error {
	const op errs.Op = "page.InsertItem"

	//tmp, err := tempFileFrom(p.file)
	//
	//if err != nil {
	//	return errs.W(op, err)
	//}
	//
	//defer os.Remove(tmp.Name())

	if err := p.incrementCount(); err != nil {
		return errs.W(op, err)
	}

	if err := p.addItemToEnd(item); err != nil {
		return errs.W(op, err)
	}

	//if err := os.Rename(tmp.Name(), p.file.Name()); err != nil {
	//	return errs.W(op, err)
	//}

	return nil
}

func (p *Page[T]) Remove() error {
	const op errs.Op = "page.Remove"

	if err := os.Remove(p.file.Name()); err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (p *Page[T]) Close() error {
	const op errs.Op = "page.Close"

	if err := p.file.Close(); err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (p *Page[T]) incrementCount() error {
	const op errs.Op = "page.incrementCount"

	var count int
	var err error

	if count, err = p.ReadCount(); err != nil {
		return errs.W(op, err)
	}

	count++

	if err = p.writeCount(count); err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (p *Page[T]) addItemToEnd(item any) error {
	const op errs.Op = "page.addItemToEnd"

	if _, err := p.file.Seek(0, io.SeekEnd); err != nil {
		return errs.W(op, err)
	}

	if err := p.encoder.Encode(item); err != nil {
		return errs.W(op, err)
	}

	return nil
}
