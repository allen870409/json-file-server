package main

import (
	"fmt"
	"io/ioutil"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"os"
	"regexp"
	"path/filepath"
)

func PUT(res http.ResponseWriter, req *http.Request) {
	osPath := filepath.FromSlash(FILE_ROOT + req.URL.Path)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprint(res, "ERROR decoding JSON - ", err)
	}

	if _, err := os.Stat(osPath); os.IsNotExist(err){
		tx, err := MyDB.Begin()
		if err != nil {
			fmt.Fprint(res, "ERROR saving to db - ", err)
			return
		}
		_, err = tx.Exec("INSERT INTO json_file (path) VALUES(?)", req.URL.Path)
		if err != nil {
			fmt.Fprint(res, "ERROR saving to db - ", err)
			tx.Rollback()
		}else{
			r := regexp.MustCompile("\\w+\\.json$")
			dirAll := r.ReplaceAllString(osPath, "")
			os.MkdirAll(dirAll, os.ModePerm)
			ioutil.WriteFile(osPath, body, os.ModeAppend)
		}
		tx.Commit()

		fmt.Fprint(res, "Create Successful!!!")
	} else {
		fmt.Fprint(res, "File exists : " + req.URL.Path)
	}
}

func LIST(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	limitStr := ""
	if len(req.Form["limit"]) > 0 {
		limitStr += " limit " + req.Form["limit"][0]
	}
	stmtOut, err := MyDB.Prepare("SELECT path FROM json_file WHERE path LIKE ?" + limitStr)
	if err != nil {
		fmt.Fprint(res, "error on prepare!")
	}
	defer stmtOut.Close()
	b := req.URL.Path + "%"
	rows, err := stmtOut.Query(b)
	if err != nil {
		fmt.Fprint(res, "error on query!")
	}
	paths := ""
	var path string
	for rows.Next(){
		err := rows.Scan(&path)
		if err != nil {
			fmt.Println(err)
		}else{
			paths += path + "\n"
		}
	}
	fmt.Fprintf(res, paths)
	rows.Close()
}

func POST(res http.ResponseWriter, req *http.Request) {
	osPath := filepath.FromSlash(FILE_ROOT + req.URL.Path)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprint(res, "ERROR decoding JSON - ", err)
	}

	if _, err := os.Stat(osPath); err == nil{
		ioutil.WriteFile(osPath, body, os.ModeAppend)
		res.WriteHeader(http.StatusCreated)
		fmt.Fprint(res, "Update Successful!!!")
	} else {
		fmt.Fprint(res, "File not exists : " + req.URL.Path)
	}
}

func DELETE(res http.ResponseWriter, req *http.Request) {

	osPath := filepath.FromSlash(FILE_ROOT + req.URL.Path)
	if _, err := os.Stat(osPath); err == nil{
		tx, err := MyDB.Begin()
		if err != nil {
			fmt.Fprint(res, "ERROR delete to db - ", err)
			return
		}
		_, err = tx.Exec("DELETE FROM json_file WHERE path=?", req.URL.Path)
		if err != nil {
			fmt.Fprint(res, "ERROR delete from db - ", err)
			tx.Rollback()
		}else{
			os.Remove(osPath)
		}
		tx.Commit()
		fmt.Fprint(res, "Delete Successful!!!")
	} else {
		fmt.Fprint(res, "File not exists : " + req.URL.Path)
	}
}
