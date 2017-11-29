package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// User Table Info
type User struct {
	ID       int64
	Name     string
	Email    string
	Password string
}

//GetAllUsers get user info
func (app *Application) GetAllUsers() ([]*User, error) {
	rows, err := app.DB.Query("SELECT UserName, Password, ID from Users")
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for rows.Next() {
		u := &User{}
		err = rows.Scan(&u.Name, &u.Password, &u.ID)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

//SaveUser Save user
func (app *Application) SaveUser(user *User) (*User, error) {
	fmt.Printf("%+v", user)
	record, err := app.DB.Exec("INSERT INTO Users (Username, Password) VALUES (?, ?)", user.Name, user.Password)
	if err != nil {
		return nil, err
	}
	i, err := record.LastInsertId()
	if err != nil {
		return nil, err
	}
	user.ID = i
	return user, err
}
