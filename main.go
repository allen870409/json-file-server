package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	. "json-file-server/controllers"
	. "json-file-server/utils"
)

func main() {
	db, err := sql.Open("mysql", "root:root@/mygo?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(100)
	defer db.Close()

	server := &Server{DB: db}

	reHandler := new(RegexpHandler)

	reHandler.HandleFunc("/todos\\.json$", "GET", server.TodoIndex)
	reHandler.HandleFunc("/todos\\.json$", "PUT", server.TodoCreate)
	reHandler.HandleFunc("/todos\\.json$", "POST", server.TodoUpdate)
	reHandler.HandleFunc("/todos/[0-9]+\\.json$", "GET", server.TodoShow)
	reHandler.HandleFunc("/todos/[0-9]+\\.json$", "DELETE", server.TodoDelete)

	fmt.Println("Starting server on port 9000")
	http.ListenAndServe(":9000", reHandler)
}
