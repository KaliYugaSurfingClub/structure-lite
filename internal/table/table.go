package table

import (
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"structure-lite/internal/errs"
	"structure-lite/internal/fsm"
	"structure-lite/internal/pages"
	"sync"
)

type Table[T any] struct {
	mtx         *sync.RWMutex
	location    string
	itemsOnPage int
	fsm         *fsm.FSM[T]
}

func New[T any](itemsOnPage int, location string) (*Table[T], error) {
	const op errs.Op = "table.New"

	location = fmt.Sprintf("%s/%s", location, ttype[T]())

	table := &Table[T]{
		itemsOnPage: itemsOnPage,
		location:    location,
		mtx:         new(sync.RWMutex),
		fsm:         fsm.New[T](),
	}

	if err := os.MkdirAll(location, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directories for %s: %w", location, err)
	}

	err := filepath.WalkDir(location, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		page, err := pages.Open[T](path)
		if err != nil {
			return fmt.Errorf("error opening page %s: %w", path, err)
		}

		log.Println("Init: Open page", path)

		count, err := page.ReadCount()
		if err != nil {
			return fmt.Errorf("error reading count of page %s: %w", path, err)
		}

		log.Println("Init: Read count", count)

		if count < itemsOnPage {
			table.fsm.Push(page, itemsOnPage-count)
			log.Println("Init: add page to fsm", path)
		}

		return nil
	})

	if err != nil {
		return nil, errs.W(op, err)
	}

	return table, nil
}

func (t *Table[T]) Put(item T) (err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	const op errs.Op = "table.Put"

	page := t.fsm.Pop()

	if page == nil {
		if page, err = t.createPage(); err != nil {
			return errs.W(op, err)
		}

		t.fsm.Push(page, t.itemsOnPage-1)
	}

	if err = page.InsertItem(item); err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (t *Table[T]) Scan(limit int, offset int) ([]T, error) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	const op errs.Op = "table.Scan"

	items := make([]T, 0, limit)
	skipped := 0

	err := filepath.WalkDir(t.location, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		page, err := pages.Open[T](path)
		if err != nil {
			return fmt.Errorf("error opening page %s: %w", path, err)
		}

		tmpItems := make([]T, 0)

		if err = page.ReadAllItems(&tmpItems); err != nil {
			return errs.W(op, err)
		}

		for _, item := range tmpItems {
			if skipped > offset {
				items = append(items, item)
			}
			skipped++
		}

		return nil
	})

	if err != nil {
		return nil, errs.W(op, err)
	}

	return items, nil
}

func (t *Table[T]) createPage() (*pages.Page[T], error) {
	const op errs.Op = "table.createPage"

	name := fmt.Sprintf("%s/%s", t.location, uuid.New().String())

	page, err := pages.Create[T](name)

	if err != nil {
		return nil, errs.W(op, err)
	}

	return page, nil
}
