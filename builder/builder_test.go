package builder_test

import (
	"reflect"
	"testing"

	"github.com/qclaogui/database/builder"
)

var App *AppService

type AppService struct {
	DB builder.Connector
	DM *builder.DatabaseManager
}

func init() {
	db, dm := builder.Run()
	App = &AppService{DB: db, DM: dm}

}

func TestWhereWithParentheses(t *testing.T) {

	fn := func() {
		App.DB.Table("users").
			Where("age", ">=", "22", "(").Where("gender", "Male").Where("house", ">=", "1", ")").
			OrWhere("age", ">=", "20", "(").Where("gender", "=", "Female", ")").
			Get()
	}

	RunDrySql(t, &TData{
		fn:       fn,
		want:     "select * from users where (age >= ? and gender = ? and house >= ?) or (age >= ? and gender = ?)",
		wantBind: []interface{}{"22", "Male", "1", "20", "Female"},
	})
}

func TestInsert(t *testing.T) {

	var insert = []map[string]string{map[string]string{
		"name":  builder.RandomString(9),
		"email": builder.RandomString(9) + "@qq.com",
	}, map[string]string{
		"name":  builder.RandomString(9),
		"email": builder.RandomString(9) + "@gmail.com",
	}}

	fn := func() {
		App.DB.Table("users").Insert(insert)
	}

	RunDrySql(t, &TData{
		fn:       fn,
		want:     "insert into users(name, email) values (?, ?), (?, ?)",
		wantBind: []interface{}{insert[0]["name"], insert[0]["email"], insert[1]["name"], insert[1]["email"]},
	})
}

func TestWhereBetween(t *testing.T) {

	fn := func() {

		App.DB.Table("users").Select().Where("name", "!=", "Go").
			WhereBetween("created_at", "2017-01-08", "2018-03-06").Get()
	}

	RunDrySql(t, &TData{
		fn:       fn,
		want:     "select * from users where name != ? and created_at between ? and ?",
		wantBind: []interface{}{"Go", "2017-01-08", "2018-03-06"},
	})
}

func TestUpdate(t *testing.T) {

	var update = map[string]string{
		"name":  builder.RandomString(9),
		"email": builder.RandomString(9) + "@qq.com"}

	fn := func() {
		App.DB.Table("users").Where("id", "1").
			WhereBetween("created_at", "2018-01-08", "2018-03-06").
			Limit(1).Update(update)
	}

	RunDrySql(t, &TData{
		fn:       fn,
		want:     "update users set name = ?, email = ? where id = ? and created_at between ? and ?",
		wantBind: []interface{}{update["name"], update["email"], "1", "2018-01-08", "2018-03-06"},
	})
}
func TestDelete(t *testing.T) {

	fn := func() {
		App.DB.Table("users").Where("id", "5").
			WhereBetween("created_at", "2018-01-08", "2018-10-06").
			Limit(1).Delete()
	}

	RunDrySql(t, &TData{
		fn:       fn,
		want:     "delete from users where id = ? and created_at between ? and ?",
		wantBind: []interface{}{"5", "2018-01-08", "2018-10-06"},
	})
}

func TestOrderBy(t *testing.T) {

	fn := func() {
		App.DB.Table("users").Select("id", "name as username").
			Where("id", ">", "2").Where("name", "Go").
			OrWhere("id", "1").Limit(2).OrderBy("id").First()
	}

	RunDrySql(t, &TData{
		fn:       fn,
		want:     "select * from users where id > ? and name = ? or id = ? order by id asc limit 1",
		wantBind: []interface{}{"2", "Go", "1"},
	})
}

func TestJoin(t *testing.T) {
	fn := func() {
		App.DB.Table("users").Where("id", ">", "2").
			WhereDay("created_at", "6").
			Join("contacts", "users.id", "contacts.user_id").
			Join("orders", "users.id", "orders.user_id").
			Select("users.*", "contacts.phone as username", "orders.price").
			Get()
	}

	RunDrySql(t, &TData{
		fn:       fn,
		want:     "select * from users inner join contacts on users.id = contacts.user_id inner join orders on users.id = orders.user_id where id > ? and day(created_at) = ?",
		wantBind: []interface{}{"2", "6"},
	})
}

type TData struct {
	fn       func()
	want     string
	wantBind []interface{}
}

func TestSelect(t *testing.T) {
	var noBind = make([]interface{}, 10)
	noBind = nil

	RunDrySql(t, &TData{
		fn:       func() { App.DB.Table("users").Select("name").Get() },
		want:     "select * from users",
		wantBind: noBind,
	})
}

func RunDrySql(t *testing.T, test *TData) {

	got := App.DB.Pretend(test.fn)

	if got[0]["query"] != test.want {
		t.Errorf("\n\x1b[91mOops🔥\x1b[39m \n✘got:\t%#v \nwant: \t%#v",
			got[0]["query"], test.want)
	}

	if !reflect.DeepEqual(got[0]["bindings"], test.wantBind) {
		t.Errorf("\n\x1b[91mOops🔥\x1b[39m \n✘got:\t%#v \nwant: \t%#v",
			got[0]["bindings"], test.wantBind)
	}
}
