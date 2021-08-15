package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Project struct {
	PrimaryKey
	Name string `json:"name"`
	Timestamps
}

func FindProject(r *http.Request) (interface{}, error) {
	id := chi.URLParam(r, "id")
	var project *Project
	Database.Find(&project, id)
	return project, nil
}

func ListProjects(_ *http.Request) (interface{}, error) {
	var projects []Project
	Database.Find(&projects)
	return projects, nil
}
