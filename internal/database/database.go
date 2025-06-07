package database

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"xorm.io/xorm"
)

func New(dbString string) (*xorm.Engine, error) {
	dbPrefix, source, found := strings.Cut(dbString, ":")
	if !found {
		return nil, errors.New("invalid database connection string format, missing ':'")
	}
	return xorm.NewEngine(dbPrefix, source)
}

type RecordIP struct {
	Id   int64
	Name string `xorm:"varchar(25) not null unique 'usr_name' comment('NickName')"`
}

type RecordTXT struct {
	Id   int64
	Name string `xorm:"varchar(25) not null unique 'usr_name' comment('NickName')"`
}
