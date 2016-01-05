package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"regexp"
)

// net/http based router
type route struct {
	pattern *regexp.Regexp
	verb    string
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, verb string, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, verb, handler})
}

func (h *RegexpHandler) HandleFunc(r string, v string, handler func(http.ResponseWriter, *http.Request)) {
	re := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{re, v, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) && route.verb == r.Method {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

// todo "Object"
type Todo struct {
	Id           int    `json:"Id"`
	Name         string `json:"Name"`
	Completed    bool `json:"Completed"`
}


// store "context" values and connections in the server struct
type Server struct {
	db *sql.DB
}

func main() {
	db, err := sql.Open("mysql", "root:root@/mygo?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(100)
	defer db.Close()

	server := &Server{db: db}

	reHandler := new(RegexpHandler)

	reHandler.HandleFunc("/todos/$", "GET", server.todoIndex)
	reHandler.HandleFunc("/todos/$", "PUT", server.todoCreate)
	reHandler.HandleFunc("/todos/$", "POST", server.todoUpdate)
	reHandler.HandleFunc("/todos/[0-9]+$", "GET", server.todoShow)
	reHandler.HandleFunc("/todos/[0-9]+$", "DELETE", server.todoDelete)

	fmt.Println("Starting server on port 9000")
	http.ListenAndServe(":9000", reHandler)
}


// Todo CRUD

func (s *Server) todoIndex(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	limitStr := ""
	if len(req.Form["limit"]) > 0 {
		limitStr += " limit " + req.Form["limit"][0]
	}
	var todos []*Todo
    fmt.Printf("%s\n", "SELECT id, name, completed FROM todo" + limitStr)
	rows, err := s.db.Query("SELECT id, name, completed FROM todo" + limitStr)
	error_check(res, err)
	for rows.Next() {
		todo := &Todo{}
		rows.Scan(&todo.Id, &todo.Name, &todo.Completed)
		todos = append(todos, todo)
	}
	rows.Close()

	jsonResponse(res, todos)
}

func (s *Server) todoCreate(res http.ResponseWriter, req *http.Request) {
	todo := &Todo{}
	body, err := ioutil.ReadAll(req.Body)
	fmt.Println("---------", string(body))
	if err != nil {
		fmt.Println("ERROR decoding JSON - ", err)
		return
	}
	err = json.Unmarshal(body, &todo)
	fmt.Println("~~~~~~~~~2~~~~~~~~~~", *todo)
	fmt.Println("===============", todo.Name, todo.Completed)
	result, err := s.db.Exec("INSERT INTO todo (name, completed) VALUES(?, ?)", todo.Name, todo.Completed)
	if err != nil {
		fmt.Println("ERROR saving to db - ", err)
	}

	Id64, err := result.LastInsertId()
	Id := int(Id64)
	todo = &Todo{Id: Id}

	s.db.
	QueryRow("SELECT id, name, completed FROM Todo WHERE id=?", todo.Id).
	Scan(&todo.Id, &todo.Name, &todo.Completed)

	jsonResponse(res, todo)
}

func (s *Server) todoShow(res http.ResponseWriter, req *http.Request) {
	r, _ := regexp.Compile(`\d+$`)

	id := r.FindString(req.URL.Path)
	fmt.Println("----------------", id)
	todo := &Todo{}
	row :=s.db.QueryRow("SELECT id, name, completed FROM todo WHERE id=?", id)
	row.Scan(&todo.Id, &todo.Name, &todo.Completed)

	jsonResponse(res, todo)
}

func (s *Server) todoUpdate(res http.ResponseWriter, req *http.Request) {
	todoParams := &Todo{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&todoParams)
	if err != nil {
		fmt.Println("ERROR decoding JSON - ", err)
		return
	}

	_, err = s.db.Exec("UPDATE Todo SET Name=?, Completed=? WHERE Id=?", todoParams.Name, todoParams.Completed, todoParams.Id)

	if err != nil {
		fmt.Println("ERROR saving to db - ", err)
	}

	todo := &Todo{Id: todoParams.Id}

	s.db.QueryRow("SELECT Id, Name, Completed FROM Todo WHERE Id=?", todo.Id).
	Scan(&todo.Id, &todo.Name, &todo.Completed)

	jsonResponse(res, todo)
}

func (s *Server) todoDelete(res http.ResponseWriter, req *http.Request) {
	r, _ := regexp.Compile(`\d+$`)
	Id := r.FindString(req.URL.Path)
	s.db.Exec("DELETE FROM Todo WHERE Id=?", Id)
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