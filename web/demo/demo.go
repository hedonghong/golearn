package demo

import (
	"encoding/json"
	"github.com/gin-gonic/gin/binding"
	"github.com/julienschmidt/httprouter"
	"golearn/web/framework"
	"net/http"
	"time"
)

type TestSearch struct {
	Id int `form:"id" json:"id"`
	Name string `form:"name" json:"name"`
	CreatedAt time.Time `form:"created_at" json:"created_at"`
}

type Test1 struct {
	Id int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	CreatedAt time.Time `form:"created_at" json:"created_at" db:"created_at"`
}

func Demo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var test1search TestSearch
	binding.Query.Bind(r, &test1search)
	tests := DemoService(test1search)
	encode := json.NewEncoder(w)
	encode.Encode(tests)
}

func DemoService(search TestSearch) []Test1 {
	var tests []Test1
	var sql string
	var where []interface{}
	sql = "select id,name,created_at from test_1 where 1=1"
	if search.Id > 0 {
		sql += " and id = ?"
		where = append(where, search.Id)
	}
	if search.Name != "" {
		sql += " and name = ?"
		where = append(where, search.Name)
	}
	err := framework.Db.Select(&tests, sql, where...)
	if err != nil {
		panic(err.Error())
	}
	return tests
}