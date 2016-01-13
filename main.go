package main

import (
	"net/http"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

const FILE_ROOT = "..\\file"
const MYSQL_USER, MYSQL_PWD, MYSQL_DB_NAME = "root", "root", "json-file-server"
const PORT = ":9000"
const PATH_REGEXP, FILE_REGEXP = "^/[\\w/]+/$", "^/[\\w/]*\\w?\\.json$"

var MyDB *sql.DB

func init(){
	var err error
	MyDB, err = sql.Open("mysql", MYSQL_USER + ":" +MYSQL_PWD + "@/" + MYSQL_DB_NAME + "?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	MyDB.SetMaxIdleConns(100)
}

func main(){
	defer MyDB.Close()
	myHandler := new(MyHandler)
	myHandler.HandleFunc("^/$", "GET", LIST)
	myHandler.HandleFunc("^/[\\w/]+/$", "GET", LIST)
	myHandler.HandleFunc("^/[\\w/]*\\w+\\.json$", "PUT", PUT)
	myHandler.HandleFunc("^/[\\w/]*\\w+\\.json$", "POST", POST)
	myHandler.HandleFunc("^/[\\w/]*\\w+\\.json$", "DELETE", DELETE)

	myHandler.HandleStatic("^/[\\w/]*\\w?\\.json$", "GET", http.FileServer(http.Dir(FILE_ROOT)))
	fmt.Println("Starting server on port : " + PORT)
	http.ListenAndServe(PORT, myHandler)
}

