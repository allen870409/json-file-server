package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"regexp"
	"database/sql"
	"strings"
	. "json-file-server/models"
)

type Server struct {
	DB *sql.DB
}

func (s *Server) TodoIndex(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	limitStr := ""
	if len(req.Form["limit"]) > 0 {
		limitStr += " limit " + req.Form["limit"][0]
	}
	var todos []*Todo
	fmt.Printf("%s\n", "SELECT id, name, completed FROM todo" + limitStr)
	rows, err := s.DB.Query("SELECT id, name, completed FROM todo" + limitStr)
	error_check(res, err)
	for rows.Next() {
		todo := &Todo{}
		rows.Scan(&todo.Id, &todo.Name, &todo.Completed)
		todos = append(todos, todo)
	}
	rows.Close()

	jsonResponse(res, todos)
}

func (s *Server) TodoCreate(res http.ResponseWriter, req *http.Request) {
	todo := &Todo{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("ERROR decoding JSON - ", err)
		return
	}
	err = json.Unmarshal(body, &todo)
	result, err := s.DB.Exec("INSERT INTO todo (name, completed) VALUES(?, ?)", todo.Name, todo.Completed)
	if err != nil {
		fmt.Println("ERROR saving to db - ", err)
	}

	Id64, err := result.LastInsertId()
	Id := int(Id64)
	todo = &Todo{Id: Id}

	s.DB.
	QueryRow("SELECT id, name, completed FROM Todo WHERE id=?", todo.Id).
	Scan(&todo.Id, &todo.Name, &todo.Completed)

	jsonResponse(res, todo)
}

func (s *Server) TodoShow(res http.ResponseWriter, req *http.Request) {
	path := strings.Replace(req.URL.Path, ".json", "", 1)
	r, _ := regexp.Compile(`\d+$`)
	id := r.FindString(path)
	todo := &Todo{}
	row :=s.DB.QueryRow("SELECT id, name, completed FROM todo WHERE id=?", id)
	row.Scan(&todo.Id, &todo.Name, &todo.Completed)

	jsonResponse(res, todo)
}

func (s *Server) TodoUpdate(res http.ResponseWriter, req *http.Request) {
	todoParams := &Todo{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&todoParams)
	if err != nil {
		fmt.Println("ERROR decoding JSON - ", err)
		return
	}

	_, err = s.DB.Exec("UPDATE Todo SET Name=?, Completed=? WHERE Id=?", todoParams.Name, todoParams.Completed, todoParams.Id)

	if err != nil {
		fmt.Println("ERROR saving to db - ", err)
	}

	todo := &Todo{Id: todoParams.Id}

	s.DB.QueryRow("SELECT Id, Name, Completed FROM Todo WHERE Id=?", todo.Id).
	Scan(&todo.Id, &todo.Name, &todo.Completed)

	jsonResponse(res, todo)
}

func (s *Server) TodoDelete(res http.ResponseWriter, req *http.Request) {
	path := strings.Replace(req.URL.Path, ".json", "", 1)
	r, _ := regexp.Compile(`\d+$`)
	id := r.FindString(path)
	s.DB.Exec("DELETE FROM Todo WHERE Id=?", id)
	res.WriteHeader(200)
}


func jsonResponse(res http.ResponseWriter, data interface{}) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")

	payload, err := json.Marshal(data)
	if error_check(res, err) {
		return
	}

	fmt.Fprintf(res, string(payload))
}

func error_check(res http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}
