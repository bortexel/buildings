package main

import (
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gopkg.in/guregu/null.v4"
)

type Project struct {
	PrimaryKey
	Name          string      `json:"name" gorm:"not null"`
	Description   null.String `json:"description"`
	Progress      int         `json:"progress" gorm:"not null"`
	LitematicaURL null.String `json:"litematica_url"`
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
	ResourceProgress ResourceProgress
}

type ResourceProgress struct {
	Done      int
	Assigned  int
	NotEnough int
	Absent    int
}

func GetResourceProgress(resources []Resource) ResourceProgress {
	resourceProgress := ResourceProgress{}

	for _, resource := range resources {
		switch resource.Status {
		case StatusDone:
			resourceProgress.Done++
		case StatusAssigned:
			resourceProgress.Assigned++
		case StatusNotEnough:
			resourceProgress.NotEnough++
		case StatusAbsent:
			resourceProgress.Absent++
		}
	}

	return resourceProgress
}

func (r *ResourceProgress) Normalize() {
	for r.GetTotal() > 100 {
		r.Absent--
	}
}

func (r *ResourceProgress) GetTotal() int {
	return r.Done + r.Assigned + r.NotEnough + r.Absent
}

func ProjectPage(r *http.Request) ProjectPageData {
	project := ProjectByID(r)
	resources := project.GetResources()
	resourceProgress := GetResourceProgress(resources)

	if len(resources) > 0 {
		resourceProgress.Done = int(math.Ceil(float64(resourceProgress.Done) / float64(len(resources)) * 100))
		resourceProgress.Assigned = int(math.Ceil(float64(resourceProgress.Assigned) / float64(len(resources)) * 100))
		resourceProgress.NotEnough = int(math.Ceil(float64(resourceProgress.NotEnough) / float64(len(resources)) * 100))
		resourceProgress.Absent = int(math.Ceil(float64(resourceProgress.Absent) / float64(len(resources)) * 100))
	} else {
		resourceProgress = ResourceProgress{}
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

func (p *Project) CreateResource(id null.String, name string, amount uint) *Resource {
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
