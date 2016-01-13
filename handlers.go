package main

import (
	"fmt"
	"io/ioutil"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"os"
	"encoding/json"
	"regexp"
	"path/filepath"
)

type ResponseJson struct{
	Status int
	Data interface{}
}

func PUT(res http.ResponseWriter, req *http.Request) {
	osPath := filepath.FromSlash(FILE_ROOT + req.URL.Path)
	body, err := ioutil.ReadAll(req.Body)
	if checkErr(res, err) {
		return
	}
	if _, err := os.Stat(osPath); os.IsNotExist(err){
		tx, err := MyDB.Begin()
		if checkErr(res, err) {
			return
		}
		_, err = tx.Exec("INSERT INTO json_file (path) VALUES(?)", req.URL.Path)
		if checkErr(res, err) {
			tx.Rollback()
			return
		}else{
			r := regexp.MustCompile("\\w+\\.json$")
			dirAll := r.ReplaceAllString(osPath, "")
			os.MkdirAll(dirAll, os.ModePerm)
			ioutil.WriteFile(osPath, body, os.ModeAppend)
		}
		tx.Commit()

		writeJson(res, &ResponseJson{Status: http.StatusOK, Data:"Create Successful!!!"})
	} else {
		body := &ResponseJson{Status: http.StatusBadRequest, Data: "File exists : " + req.URL.Path}
		writeJson(res, body)
	}
}

func LIST(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	limitStr := ""
	if len(req.Form["limit"]) > 0 {
		limitStr += " limit " + req.Form["limit"][0]
	}
	stmtOut, err := MyDB.Prepare("SELECT path FROM json_file WHERE path LIKE ?" + limitStr)
	if checkErr(res, err) {
		return
	}
	defer stmtOut.Close()
	b := req.URL.Path + "%"
	rows, err := stmtOut.Query(b)
	if checkErr(res, err) {
		return
	}
	var paths []string
	var path string
	for rows.Next(){
		err := rows.Scan(&path)
		if checkErr(res, err) {
			return
		}else{
			paths = append(paths, path)
		}
	}
	body := &ResponseJson{Status:http.StatusOK, Data: paths}
	writeJson(res, body)
	rows.Close()
}

func POST(res http.ResponseWriter, req *http.Request) {
	osPath := filepath.FromSlash(FILE_ROOT + req.URL.Path)
	body, err := ioutil.ReadAll(req.Body)
	if checkErr(res, err) {
		return
	}
	_, e := os.Stat(osPath)
	if checkErr(res, e) {
		return
	}else{
		ioutil.WriteFile(osPath, body, os.ModeAppend)
		res.WriteHeader(http.StatusCreated)
		writeJson(res, &ResponseJson{Status: http.StatusOK, Data:"Update Successful!!!"})
	}
}

func DELETE(res http.ResponseWriter, req *http.Request) {

	osPath := filepath.FromSlash(FILE_ROOT + req.URL.Path)
	if _, err := os.Stat(osPath); err == nil{
		tx, err := MyDB.Begin()
		if checkErr(res, err) {
			return
		}
		_, err = tx.Exec("DELETE FROM json_file WHERE path=?", req.URL.Path)
		if checkErr(res, err) {
			tx.Rollback()
			return
		}else{
			os.Remove(osPath)
		}
		tx.Commit()
		writeJson(res, &ResponseJson{Status: http.StatusOK, Data:"Delete Successful!!!"})
	} else {
		writeJson(res, &ResponseJson{Status: http.StatusBadRequest, Data:"File not exists : " + req.URL.Path})
	}
}

func writeJson(res http.ResponseWriter, data interface{}) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	payload, err := json.Marshal(data)
	if checkErr(res, err) {
		return
	}
	fmt.Fprintf(res, string(payload))
}

func checkErr(res http.ResponseWriter, err error) bool {
	if err != nil {
		result := &ResponseJson{Status:http.StatusInternalServerError, Data: err.Error()}
		payload, _ := json.Marshal(result)
		fmt.Fprintf(res, string(payload))
		return true
	}
	return false
}