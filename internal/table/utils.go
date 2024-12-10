package table

import (
	"reflect"
)

func ttype[T any]() string {
	return reflect.TypeOf((*T)(nil)).Elem().String()
}
