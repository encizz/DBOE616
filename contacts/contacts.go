package contacts

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/kataras/golog"
)

var (
	ConnSQL *pgx.ConnConfig
	CtxSQL  context.Context
)

type Contact struct {
	ID     int    `json:"id"`
	Name   Name   `json:"name"`
	Phone  Phone  `json:"phone"`
	Web    Web    `json:"web_site"`
	Adress Adress `json:"adress"`
	Mail   Mail   `json:"mail"`
}
type Name struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type Phone struct {
	ID    int    `json:"id"`
	Phone string `json:"phone"`
}
type Web struct {
	ID  int    `json:"id"`
	Web string `json:"web"`
}
type Adress struct {
	ID     int    `json:"id"`
	Adress string `json:"adress"`
}
type Mail struct {
	ID   int    `json:"id"`
	Mail string `json:"mail"`
}

func (contact Contact) InsertContact() (success bool) {

	query := "insert into contact (id, name,adress,phone,web_site,mail) " +
		"values ($1,$2,$3,$4,$5,$6) "

	sql, err := pgx.ConnectConfig(CtxSQL, ConnSQL)
	if err != nil {
		return
	}
	defer sql.Close(CtxSQL)

	_, err = sql.Exec(CtxSQL, query,
		contact.ID,
		contact.Name,
		contact.Adress,
		contact.Phone,
		contact.Web,
	)

	if err != nil {
		golog.Error(query)
		fmt.Println(err.Error())
	} else {
		success = true
	}
	return
}
