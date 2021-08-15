package main

type Project struct {
	PrimaryKey
	Name string `json:"name"`
	Timestamps
}
