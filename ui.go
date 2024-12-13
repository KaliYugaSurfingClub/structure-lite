package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"log"
	"strconv"
	"strings"
	"time"
)

type UserTableApp struct {
	App          fyne.App
	Window       fyne.Window
	UserList     *widget.List
	NameInput    *widget.Entry
	EmailInput   *widget.Entry
	AddressInput *widget.Entry
	TagsInput    *widget.Entry
	AgeInput     *widget.Entry
	LimitInput   *widget.Entry // Ввод для лимита
	OffsetInput  *widget.Entry // Ввод для оффсета
	AddButton    *widget.Button
	DeleteButton *widget.Button
	UpdateButton *widget.Button // Кнопка для обновления списка
	UserManager  *UserManager
	UserData     []string
}

func NewUserTableApp(manager *UserManager) *UserTableApp {
	a := app.New()
	w := a.NewWindow("User Table")

	// Устанавливаем окно на всю ширину экрана
	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(false) // Окно будет изменять размеры

	userData := []string{}

	// Список пользователей
	userList := widget.NewList(
		func() int { return len(userData) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(userData[i])
		},
	)

	// Поля ввода для добавления нового пользователя
	nameInput := widget.NewEntry()
	nameInput.SetPlaceHolder("Enter user name")

	emailInput := widget.NewEntry()
	emailInput.SetPlaceHolder("Enter user email")

	addressInput := widget.NewEntry()
	addressInput.SetPlaceHolder("Enter user address")

	tagsInput := widget.NewEntry()
	tagsInput.SetPlaceHolder("Enter tags (comma-separated)")

	ageInput := widget.NewEntry()
	ageInput.SetPlaceHolder("Enter age (number)")

	// Поля для лимита и оффсета
	limitInput := widget.NewEntry()
	limitInput.SetPlaceHolder("Enter limit (number)")

	offsetInput := widget.NewEntry()
	offsetInput.SetPlaceHolder("Enter offset (number)")

	// Кнопка для добавления нового пользователя
	addButton := widget.NewButton("Add User", func() {
		name := nameInput.Text
		email := emailInput.Text
		address := addressInput.Text
		tags := strings.Split(tagsInput.Text, ",")
		age, err := strconv.Atoi(ageInput.Text)
		if err != nil {
			log.Println("Invalid age format")
			return
		}

		if name == "" || email == "" {
			log.Println("Name and Email are required fields")
			return
		}

		user := User{
			Name:      name,
			Email:     email,
			Address:   address,
			Tags:      tags,
			Age:       age,
			CreatedAt: time.Now(),
		}

		err = manager.AddUser(user)
		if err != nil {
			log.Printf("Error adding user: %v", err)
			return
		}

		refreshUserList(userList, manager, &userData, 100, 0)
		nameInput.SetText("")
		emailInput.SetText("")
		addressInput.SetText("")
		tagsInput.SetText("")
		ageInput.SetText("")
	})

	// Кнопка для удаления пользователя по имени
	deleteButton := widget.NewButton("Delete User by Name", func() {
		name := nameInput.Text
		if name == "" {
			log.Println("Name is required to delete a user")
			return
		}

		err := manager.DeleteUserByName(name)
		if err != nil {
			log.Printf("Error deleting user: %v", err)
			return
		}

		refreshUserList(userList, manager, &userData, 100, 0)
		nameInput.SetText("")
	})

	// Кнопка для обновления списка с учётом лимита и оффсета
	updateButton := widget.NewButton("Update List", func() {
		limit, err := strconv.Atoi(limitInput.Text)
		if err != nil || limit <= 0 {
			log.Println("Invalid limit")
			return
		}

		offset, err := strconv.Atoi(offsetInput.Text)
		if err != nil || offset < 0 {
			log.Println("Invalid offset")
			return
		}

		refreshUserList(userList, manager, &userData, limit, offset)
	})

	// Формы для ввода данных
	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Name", nameInput),
			widget.NewFormItem("Email", emailInput),
			widget.NewFormItem("Address", addressInput),
			widget.NewFormItem("Tags", tagsInput),
			widget.NewFormItem("Age", ageInput),
		),
		container.NewHBox(addButton, deleteButton),
		widget.NewForm(
			widget.NewFormItem("Limit", limitInput),
			widget.NewFormItem("Offset", offsetInput),
		),
		updateButton, // Кнопка для обновления списка
	)

	content := container.NewVSplit(
		userList, // Вставляем пустой контейнер для растяжения
		form,
	)

	// Настройка контента окна
	w.SetContent(content)

	return &UserTableApp{
		App:          a,
		Window:       w,
		UserList:     userList,
		NameInput:    nameInput,
		EmailInput:   emailInput,
		AddressInput: addressInput,
		TagsInput:    tagsInput,
		AgeInput:     ageInput,
		LimitInput:   limitInput,
		OffsetInput:  offsetInput,
		AddButton:    addButton,
		DeleteButton: deleteButton,
		UpdateButton: updateButton,
		UserManager:  manager,
		UserData:     userData,
	}
}

// Функция для обновления списка пользователей с учётом лимита и оффсета
func refreshUserList(userList *widget.List, manager *UserManager, userData *[]string, limit, offset int) {
	users, err := manager.ListUsers(limit, offset)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		return
	}

	*userData = []string{}
	for _, user := range users {
		*userData = append(*userData, formatUser(user))
	}
	userList.Refresh()
}

// Функция для форматирования информации о пользователе
func formatUser(user User) string {
	return "Name: " + user.Name + ", Email: " + user.Email + ", Address: " + user.Address +
		", Tags: " + strings.Join(user.Tags, ",") + ", Age: " + strconv.Itoa(user.Age) +
		", CreatedAt: " + user.CreatedAt.Format("2006-01-02")
}

// Метод для запуска приложения
func (app *UserTableApp) Run() {
	refreshUserList(app.UserList, app.UserManager, &app.UserData, 100, 0)
	app.Window.ShowAndRun()
}
