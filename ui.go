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
	AddButton    *widget.Button
	DeleteButton *widget.Button
	UserManager  *UserManager
	UserData     []string
}

func NewUserTableApp(manager *UserManager) *UserTableApp {
	a := app.New()
	w := a.NewWindow("User Table")
	w.Resize(fyne.NewSize(800, 600))

	userData := []string{}

	userList := widget.NewList(
		func() int { return len(userData) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(userData[i])
		},
	)

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

		refreshUserList(userList, manager, &userData)
		nameInput.SetText("")
		emailInput.SetText("")
		addressInput.SetText("")
		tagsInput.SetText("")
		ageInput.SetText("")
	})

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

		refreshUserList(userList, manager, &userData)
		nameInput.SetText("")
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Name", nameInput),
			widget.NewFormItem("Email", emailInput),
			widget.NewFormItem("Address", addressInput),
			widget.NewFormItem("Tags", tagsInput),
			widget.NewFormItem("Age", ageInput),
		),
		container.NewHBox(addButton, deleteButton),
	)

	content := container.NewHSplit(
		container.NewVBox(widget.NewLabel("User List"), userList),
		form,
	)
	content.SetOffset(0.7)

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
		AddButton:    addButton,
		DeleteButton: deleteButton,
		UserManager:  manager,
		UserData:     userData,
	}
}

func refreshUserList(userList *widget.List, manager *UserManager, userData *[]string) {
	users, err := manager.ListUsers(100, 0)
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

func formatUser(user User) string {
	return "Name: " + user.Name + ", Email: " + user.Email + ", Address: " + user.Address +
		", Tags: " + strings.Join(user.Tags, ",") + ", Age: " + strconv.Itoa(user.Age) +
		", CreatedAt: " + user.CreatedAt.Format("2006-01-02")
}

func (app *UserTableApp) Run() {
	refreshUserList(app.UserList, app.UserManager, &app.UserData)
	app.Window.ShowAndRun()
}
