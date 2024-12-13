package tables

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"slices"
	"structure-lite/internal/errs"
	"structure-lite/internal/fsm"
	"structure-lite/internal/pages"
	"sync"
)

type Table[T any] struct {
	mtx         *sync.RWMutex
	fsm         *fsm.FSM[T]
	location    string
	itemsOnPage int
	log         logger
}

func New[T any](itemsOnPage int, location string, log logger) (*Table[T], error) {
	const op errs.Op = "t.New"

	location = fmt.Sprintf("%s/%s", location, ttype[T]())

	t := &Table[T]{
		itemsOnPage: itemsOnPage,
		location:    location,
		mtx:         new(sync.RWMutex),
		fsm:         fsm.New[T](),
		log:         log,
	}

	if err := os.MkdirAll(location, 0666); err != nil {
		return nil, fmt.Errorf("failed to create directories for %s: %w", location, err)
	}

	err := forPages[T](location, func(page *pages.Page[T]) error {
		t.info("%s: Open page: %s", op, page.Name())

		count, err := page.ReadCount()
		if err != nil {
			return err
		}

		t.info("%s: Read count: %d of page: %s", op, count, page.Name())

		if count < itemsOnPage {
			t.fsm.Push(page, itemsOnPage-count)
		}

		return nil
	})

	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return t, nil
}

func (t *Table[T]) Insert(item T) (err error) {
	const op errs.Op = "table.Insert"

	t.mtx.Lock()
	defer t.mtx.Unlock()

	page := t.fsm.Pop()

	if page == nil {
		if page, err = pages.Create[T](t.newPageName()); err != nil {
			return errs.Wrap(op, err)
		}

		t.info("%s: not find free space, create new page and push it in fsm: %s", op, page.Name())

		t.fsm.Push(page, t.itemsOnPage-1)
	}

	if err = page.InsertItem(item); err != nil {
		return errs.Wrap(op, err)
	}

	t.info("%s: insert item: %+v to page: %s", op, item, page.Name())

	return nil
}

func (t *Table[T]) Scan(limit int, offset int) ([]T, error) {
	const op errs.Op = "table.Scan"

	items, err := t.ScanFunc(limit, offset, func(T) bool { return true })
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return items, nil
}

func (t *Table[T]) ScanFunc(limit int, offset int, filter func(T) bool) ([]T, error) {
	const op errs.Op = "table.Scan"

	t.mtx.RLock()
	defer t.mtx.RUnlock()

	items := make([]T, 0, limit)
	skipped := 0

	err := forPages(t.location, func(page *pages.Page[T]) error {
		tmpItems := make([]T, 0, t.itemsOnPage)

		if err := page.ReadAllItems(&tmpItems); err != nil {
			return err
		}

		t.info("%s: Read items: %+v of page: %s", op, tmpItems, t.location)

		filtered := slices.DeleteFunc(tmpItems, func(item T) bool {
			return !filter(item)
		})

		for _, item := range filtered {
			if len(items) == limit {
				break
			}

			if skipped >= offset {
				items = append(items, item)
			}

			skipped++
		}

		return nil
	})

	if err != nil {
		return nil, errs.Wrap(op, err)
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
			return errs.Wrap(op, err)
		}

		afterDelete := slices.DeleteFunc(items, delFunc)

		if len(items) == len(afterDelete) {
			t.info("%s: open and not find items for delete, not change page: %s", op, page.Name())
			return nil
		}

		if err := page.Trunc(); err != nil {
			return errs.Wrap(op, err)
		}

		for _, item := range afterDelete {
			if err := page.InsertItem(item); err != nil {
				return errs.Wrap(op, err)
			}
		}

		t.info("%s: open and delete items from page: %s", op, page.Name())

		return nil
	})

	if err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (t *Table[T]) newPageName() string {
	return fmt.Sprintf("%s/%s", t.location, uuid.New().String())
}

func (t *Table[T]) info(format string, args ...any) {
	t.log.Info(fmt.Sprintf(format, args...))
}
