package main

import (
	"fmt"
	time "time"

	_ "github.com/go-sql-driver/mysql"
)

// Groups Table
type CreateAccount struct {
	ID              int64
	FullName        string
	Sex             string
	Email           string
	Date            time.Time
	School          string
	Password        string
	ConfirmPassword string
}

//Date := time.Now().Format(YYYY-MM-DD)
//GetAllAccount get account info
func (app *Application) GetAllAccount() ([]*CreateAccount, error) {
	rows, err := app.DB.Query("SELECT ID, FullName,Sex,Email,Date,School,Password,ConfirmPassword FROM CreateAccount")
	if err != nil {
		return nil, err
	}

	create := []*CreateAccount{}
	for rows.Next() {
		c := &CreateAccount{}
		err = rows.Scan(&c.ID, &c.FullName, &c.Sex, &c.Email, &c.Date, &c.School, &c.Password, &c.ConfirmPassword)
		if err != nil {
			return nil, err
		}
		create = append(create, c)
	}
	return create, nil
}

//SaveAccount Save Account
func (app *Application) SaveAccount(CreateAccount *CreateAccount) (*CreateAccount, error) {
	fmt.Printf("%+v", CreateAccount)
	record, err := app.DB.Exec("INSERT INTO CreateAccount (FullName,Sex,Email,Date,School,Password,ConfirmPassword) VALUES (?, ?, ?, ? ,?, ?, ? )", &CreateAccount.FullName, &CreateAccount.Sex, &CreateAccount.Email, &CreateAccount.Date, &CreateAccount.School, &CreateAccount.Password, &CreateAccount.ConfirmPassword)
	if err != nil {
		return nil, err
	}
	i, err := record.LastInsertId()
	if err != nil {
		return nil, err
	}
	CreateAccount.ID = i
	return CreateAccount, err
}
