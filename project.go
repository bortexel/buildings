package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gopkg.in/guregu/null.v4"
)

type Project struct {
	PrimaryKey
	Name        string      `json:"name" gorm:"not null"`
	Description null.String `json:"description"`
	Progress    int         `json:"progress" gorm:"not null"`
	Timestamps
}

func ProjectByID(r *http.Request) *Project {
	id := chi.URLParam(r, "id")
	var project *Project
	Database.Find(&project, id)
	return project
}

type ProjectPageData struct {
	Project          *Project
	Resources        []Resource
	ResourceProgress int
}

func ProjectPage(r *http.Request) ProjectPageData {
	project := ProjectByID(r)
	resources := project.GetResources()

	completedResources := 0
	for _, resource := range resources {
		if resource.Status == StatusCompleted {
			completedResources++
		}
	}

	resourceProgress := 0
	if len(resources) > 0 {
		resourceProgress = int(float64(completedResources) / float64(len(resources)) * 100)
	}

	return ProjectPageData{
		Project:          project,
		Resources:        resources,
		ResourceProgress: resourceProgress,
	}
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

func (p *Project) GetResources() []Resource {
	var resources []Resource
	Database.Find(&resources, "project_id = ?", p.ID)
	return resources
}

func (p Project) GetDescription() string {
	return p.Description.String
}
