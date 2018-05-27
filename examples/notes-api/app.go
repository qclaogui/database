package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/qclaogui/database/builder"
)

// CREATE TABLE `notes` (
//  `title` varchar(255) DEFAULT NULL,
//  `body` text,
//  `id` int(11) NOT NULL AUTO_INCREMENT,
//  PRIMARY KEY (`id`)
// ) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4;

// App application name
var App *AppService

// AppService service
type AppService struct {
	DB     builder.Connector
	DM     *builder.DatabaseManager
	router *httprouter.Router
}

// NewAppService new service
func NewAppService() *AppService {

	// all done
	db, dm := builder.Run("/absolute/path/to/database.yml")

	return &AppService{
		DB:     db,
		DM:     dm,
		router: httprouter.New(),
	}
}

// postman api collections
// https://www.getpostman.com/collections/1a1777b69c8b61d8c180

// AppService routes
func (s *AppService) routes() {
	// welcome page
	s.router.GET("/", s.welcome)
	s.router.GET("/notes", s.GetAll)
	s.router.POST("/notes", s.Create)
	s.router.GET("/notes/:id", s.GetOne)
	s.router.PUT("/notes/:id", s.Update)
	s.router.DELETE("/notes/:id", s.Destroy)
}

type resBody struct {
	ErrCode int         `json:"err_code" xml:"err_code"`
	Data    interface{} `json:"data" xml:"data"`
	ErrMsg  string      `json:"err_msg" xml:"err_msg"`
}

func resOK(data interface{}) *resBody {
	return &resBody{ErrCode: 0, Data: data, ErrMsg: "ok"}
}

func resNotOK(data interface{}) *resBody {
	return &resBody{ErrCode: 40001, Data: data, ErrMsg: "Oops! somthing wrong"}
}

func toJSON(w http.ResponseWriter, result *resBody) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// welcome msg
func (s *AppService) welcome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	toJSON(w, resOK("Welcome gopher!"))
}

// Create CURD Create
func (s *AppService) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var notes []map[string]string

	note := map[string]string{
		"title": r.PostFormValue("title"),
		"body":  r.PostFormValue("body"),
	}

	notes = append(notes, note)

	res := s.DB.Table("notes").Insert(notes)

	if res < 1 {
		toJSON(w, resNotOK(false))
	} else {
		toJSON(w, resOK(true))
	}
}

// Update CURD Update
func (s *AppService) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	note := map[string]string{
		"title": r.PostFormValue("title"),
		"body":  r.PostFormValue("body"),
	}

	res := s.DB.Table("notes").Where("id", ps.ByName("id")).Update(note)

	if res < 1 {
		toJSON(w, resNotOK(false))
	} else {
		toJSON(w, resOK(true))
	}
}

// GetOneByTitle CURD Retrieve
func (s *AppService) GetOne(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	res := s.DB.Table("notes").Where("id", ps.ByName("id")).First()

	if len(res) < 1 {
		toJSON(w, resNotOK(false))
	} else {
		toJSON(w, resOK(res))
	}
}

// Destroy CURD Delete
func (s *AppService) Destroy(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	res := s.DB.Table("notes").Where("id", ps.ByName("id")).Delete()

	if res < 1 {
		toJSON(w, resNotOK(false))
	} else {
		toJSON(w, resOK(true))
	}
}

// GetAll get all notes
func (s *AppService) GetAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	res := s.DB.Table("notes").Limit(1000).Get()

	if len(res) < 1 {
		toJSON(w, resNotOK("no notes"))
	} else {
		toJSON(w, resOK(res))
	}
}

func main() {

	// create Service
	server := NewAppService()

	// add routes
	server.routes()

	// Run
	log.Println("server run at http://localhost:8088")
	log.Fatal(http.ListenAndServe(":8088", server.router))
}
