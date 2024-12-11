package tables

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"slices"
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

	err := forPages[T](location, func(page *pages.Page[T]) error {
		log.Printf("Init: Open page: %s\n", page.Name())

		count, err := page.ReadCount()
		if err != nil {
			return fmt.Errorf("error reading count of page %s: %w", page.Name(), err)
		}

		log.Printf("Init: Read count %d from %s\n", count, page.Name())

		if count < itemsOnPage {
			table.fsm.Push(page, itemsOnPage-count)
		}

		return nil
	})

	if err != nil {
		return nil, errs.W(op, err)
	}

	return table, nil
}

func (t *Table[T]) Insert(item T) (err error) {
	const op errs.Op = "table.Insert"

	t.mtx.Lock()
	defer t.mtx.Unlock()

	page := t.fsm.Pop()

	if page == nil {
		if page, err = pages.Create[T](t.newPageName()); err != nil {
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
	const op errs.Op = "table.Scan"

	t.mtx.RLock()
	defer t.mtx.RUnlock()

	items := make([]T, 0, limit)
	skipped := 0

	err := forPages(t.location, func(page *pages.Page[T]) error {
		tmpItems := make([]T, 0, t.itemsOnPage)

		if err := page.ReadAllItems(&tmpItems); err != nil {
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

func (t *Table[T]) Delete(delFunc func(T) bool) error {
	const op errs.Op = "table.Delete"

	t.mtx.Lock()
	defer t.mtx.Unlock()

	err := forPages(t.location, func(page *pages.Page[T]) error {
		items := make([]T, 0, t.itemsOnPage)

		if err := page.ReadAllItems(&items); err != nil {
			return errs.W(op, err)
		}

		slices.DeleteFunc(items, delFunc)

		newPage, err := pages.CreateFromItems(t.newPageName(), items)
		if err != nil {
			return errs.W(op, err)
		}

		defer pages.DeleteIfErr(err, newPage)

		err = page.Delete()
		if err != nil {
			return errs.W(op, err)
		}

		t.fsm.Push(newPage, len(items))

		return nil
	})

	if err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (t *Table[T]) newPageName() string {
	return fmt.Sprintf("%s/%s", t.location, uuid.New().String())
}
