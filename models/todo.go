package models

type Todo struct {
	Id           int    `json:"Id"`
	Name         string `json:"Name"`
	Completed    bool `json:"Completed"`
}
