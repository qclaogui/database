package builder_test

import (
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
func checkSelectSql(tp string, b *builder.Builder, t *testing.T) {
	// log.Printf("\x1b[92m Builder addr: \x1b[39m%p", b)
	switch tp {
	case "select":
		got := b.Value("id")
		if got != "1" {
			t.Errorf("\n\x1b[91mOops🔥\x1b[39m \n✘got:\t%#v \nwant: \t%#v",
				got, "1")
		}
	case "exist":
		got := b.Exists()
		if !got {
			t.Errorf("\n\x1b[91mOops🔥\x1b[39m \n✘got:\t%#v \nwant: \t%#v",
				got, true)
		}
	case "insert":
		var insert = []map[string]string{map[string]string{
			"name":  builder.RandomString(9),
			"email": builder.RandomString(9) + "@qq.com",
		}, map[string]string{
			"name":  builder.RandomString(9),
			"email": builder.RandomString(9) + "@gmail.com",
		}}
		got := b.Insert(insert)
		if got != 2 {
			t.Errorf("\n\x1b[91mOops🔥\x1b[39m \n✘got:\t%#v \nwant: \t%#v",
				got, 2)
		}
	case "update":
		update := map[string]string{
			"name":  builder.RandomString(9),
			"email": builder.RandomString(9) + "@qq.com"}
		got := b.Update(update)
		if got != 1 {
			t.Errorf("\n\x1b[91mOops🔥\x1b[39m \n✘got:\t%#v \nwant: \t%#v",
				got, 1)
		}
	case "delete":
		got := b.Delete()
		if got != 1 {
			t.Errorf("\n\x1b[91mOops🔥\x1b[39m \n✘got:\t%#v \nwant: \t%#v",
				got, 1)
		}
	}
}

func TestWhereExist(t *testing.T) {
	b := App.DB.Table("users").Select("name").Where("id", ">", "1").
		WhereTime("created_at", "=", "13:25:46")
	checkSelectSql("exist", b, t)
}

func TestWhereBetween(t *testing.T) {
	b := App.DB.Table("users").Select().Where("name", "!=", "Go").
		WhereBetween("created_at", "2017-01-08", "2018-03-06")
	checkSelectSql("select", b, t)
}

func TestOrderBy(t *testing.T) {
	b := App.DB.Table("users").Select("id", "name as username").
		Where("id", ">", "2").Where("name", "Go").
		OrWhere("id", "1").Limit(2).OrderBy("id")

	checkSelectSql("select", b, t)
}

func TestJoin(t *testing.T) {
	// b := App.DB.Table("users").Debug().Where("id", ">", "2").
	// 	WhereDay("created_at", "6").
	// 	Join("contacts", "users.id", "contacts.user_id").
	// 	Join("orders", "users.id", "orders.user_id").
	// 	Select("users.*", "contacts.phone as username", "orders.price")
	//
	// checkSelectSql("select", builder,  t)
}

func TestInsert(t *testing.T) {
	b := App.DB.Table("users")

	checkSelectSql("insert", b, t)
}

func TestUpdate(t *testing.T) {
	b := App.DB.Table("users").Where("id", "1").
		WhereBetween("created_at", "2018-01-08", "2018-03-06").
		Limit(1)

	checkSelectSql("update", b, t)
}
func TestDelete(t *testing.T) {
	b := App.DB.Table("users").Where("id", "5").
		WhereBetween("created_at", "2018-01-08", "2018-10-06").
		Limit(1)

	checkSelectSql("delete", b, t)
}

func TestSelect(t *testing.T) {
	b := App.DB.Table("users").Select("name")
	checkSelectSql("select", b, t)
}
