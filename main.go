package main

import (
	"fmt"
	"structure-lite/internal/table"
	"time"
)

type User struct {
	Name      string
	Age       int
	CreatedAt time.Time
	PhotoURLs []string
}

func init() {

}

func main() {
	//file, err := os.OpenFile("./data/main.User/37558764-3d58-4031-8cb2-badd1a0089a2", os.O_RDWR, os.ModePerm)
	//if err != nil {
	//	panic(err)
	//}
	//file.Seek(4, io.SeekStart)
	//
	//for {
	//	item := new(User)
	//	err := gob.NewDecoder(file).Decode(item)
	//
	//	if err == io.EOF {
	//		break
	//	}
	//
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	fmt.Println(item)
	//}

	t, err := table.New[User](4, "./data")
	if err != nil {
		panic(err)
	}

	err = t.Put(User{
		Name:      "123",
		Age:       123,
		PhotoURLs: []string{"123", "12s", "213"},
	})
	if err != nil {
		panic(err)
	}
	err = t.Put(User{
		Name:      "asd",
		Age:       213,
		PhotoURLs: []string{"asdsad", "asdasd", "aadas"},
	})
	if err != nil {
		panic(err)
	}
	err = t.Put(User{
		Name:      "87876tyty",
		Age:       7,
		PhotoURLs: []string{"saccxx", "xczxczzcx", "asdsadadsa"},
	})
	if err != nil {
		panic(err)
	}
	err = t.Put(User{
		Name:      "777",
		Age:       123,
		PhotoURLs: []string{"ssdsad", "123123", "213312"},
	})
	if err != nil {
		panic(err)
	}

	item, err := t.Scan(3, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", item)

	err = t.Put(User{
		Name:      "87876tyty",
		Age:       7,
		PhotoURLs: []string{"saccxx", "xczxczzcx", "asdsadadsa"},
	})
	if err != nil {
		panic(err)
	}
	err = t.Put(User{
		Name:      "777",
		Age:       123,
		PhotoURLs: []string{"ssdsad", "123123", "213312"},
	})
	if err != nil {
		panic(err)
	}

	item, err = t.Scan(1, 4)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", item)
}
