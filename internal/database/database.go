package database

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"net/netip"
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
	Name string `xorm:"varchar(255) not null 'name'"`
	Addr dbAddr `xorm:"varchar(40) not null unique 'usr_name' comment('NickName')"` // 40 length is enough for a full IPv6 address 0011:2233:4455:6677:8899:aabb:ccdd:eeff
}

type RecordTXT struct {
	Id    int64
	Name  string `xorm:"varchar(255) not null 'name'"`
	Value string `xorm:"varchar(4096) not null 'value'"`
}

type dbAddr netip.Addr

func (d *dbAddr) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("invalid IP address")
	}
	addr, err := netip.ParseAddr(str)
	if err != nil {
		return err
	}
	*d = dbAddr(addr)
	return nil
}

func (d dbAddr) Value() (driver.Value, error) {
	return netip.Addr(d).String(), nil
}

var _ driver.Valuer = dbAddr{}
var _ sql.Scanner = (*dbAddr)(nil)
