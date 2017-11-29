package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Groups Table
type Groups struct {
	ID        int64
	Title     string
	Subject   string
	MaxPeople int64
	User      int64
	GroupLead string
}

//GetAllGroups get group info
func (app *Application) GetAllGroups() ([]*Groups, error) {
	rows, err := app.DB.Query("SELECT Title,Subject,MaxPeople,user,GroupLead, ID from Groups")
	if err != nil {
		return nil, err
	}

	groups := []*Groups{}
	for rows.Next() {
		g := &Groups{}
		err = rows.Scan(&g.Title, &g.Subject, &g.MaxPeople, &g.User, &g.GroupLead, &g.ID)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

//SaveGroup Save group
func (app *Application) SaveGroup(Groups *Groups) (*Groups, error) {
	fmt.Printf("%+v", Groups)
	record, err := app.DB.Exec("INSERT INTO Groups (Title,Subject,MaxPeople,user,GroupLead) VALUES (?, ?, ?, ? ,? )", &Groups.Title, &Groups.Subject, &Groups.MaxPeople, &Groups.User, &Groups.GroupLead)
	if err != nil {
		return nil, err
	}
	i, err := record.LastInsertId()
	if err != nil {
		return nil, err
	}
	Groups.ID = i
	return Groups, err
}
