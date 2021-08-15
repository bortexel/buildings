package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Project struct {
	PrimaryKey
	Name        string `json:"name"`
	Description string `json:"description"`
	Timestamps
}

func ProjectByID(r *http.Request) *Project {
	id := chi.URLParam(r, "id")
	var project *Project
	Database.Find(&project, id)
	return project
}

func FindProject(r *http.Request) (interface{}, error) {
	return ProjectByID(r), nil
}

func AllProjects() []Project {
	var projects []Project
	Database.Find(&projects)
	return projects
}

func ListProjects(_ *http.Request) (interface{}, error) {
	return AllProjects(), nil
}

func (p *Project) CreateResource(id string, name string, amount uint) *Resource {
	resource := &Resource{
		MinecraftID: id,
		Name:        name,
		Amount:      amount,
		ProjectID:   p.ID,
	}

	Database.Save(resource)
	return resource
}
