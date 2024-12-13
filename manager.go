package main

import (
	"strings"
	"structure-lite/internal/tables"
)

type UserManager struct {
	table *tables.Table[User]
}

func NewUserManager(table *tables.Table[User]) *UserManager {
	return &UserManager{table: table}
}

func (u *UserManager) AddUser(user User) error {
	return u.table.Insert(user)
}

func (u *UserManager) DeleteUserByName(name string) error {
	return u.table.Delete(func(user User) bool {
		return strings.EqualFold(user.Name, name)
	})
}

func (u *UserManager) ListUsers(limit, offset int) ([]User, error) {
	return u.table.Scan(limit, offset)
}
