package framework

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

var Db *DB

type DB struct {
	dbConn *sqlx.DB
}


//这里重新封装sqlx参数
type MySqlx interface {
	Select(dest interface{}, query string, args ...interface{}) error
}

func init()  {
	dns  := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true", "root", "123456", "127.0.0.1", "test")
	conn := sqlx.MustConnect("mysql", dns)
	Db = &DB {
		dbConn: conn,
	}
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(10)
}

//日志可以单独出来搞成配置
func (d *DB) Select(dest interface{}, query string, args ...interface{}) error {
	defer func(start time.Time) {
		end := time.Since(start)
		fmt.Println(query)
		fmt.Println(args)
		fmt.Println(end)
	}(time.Now())
	return d.dbConn.Select(dest, query, args...)
}

func (d *DB) Close() error {
	return d.dbConn.Close()
}
