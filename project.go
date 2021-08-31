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
	Done      uint
	Assigned  uint
	NotEnough uint
	Absent    uint
}

func GetResourceProgress(resources []Resource) ResourceProgress {
	resourceProgress := ResourceProgress{}

	for _, resource := range resources {
		switch resource.Status {
		case StatusDone:
			resourceProgress.Done += resource.Amount
		case StatusAssigned:
			resourceProgress.Assigned += resource.Amount
		case StatusNotEnough:
			resourceProgress.NotEnough += resource.Amount
		case StatusAbsent:
			resourceProgress.Absent += resource.Amount
		}
	}

	return resourceProgress
}

func (r *ResourceProgress) Normalize() {
	for r.GetTotal() > 100 {
		if r.Absent > 0 {
			r.Absent--
			continue
		}

		if r.Assigned > 0 {
			r.Assigned--
			continue
		}

		if r.NotEnough > 0 {
			r.NotEnough--
			continue
		}

		if r.Done > 0 {
			r.Done--
			continue
		}
	}
}

func (r *ResourceProgress) GetTotal() uint {
	return r.Done + r.Assigned + r.NotEnough + r.Absent
}

func ProjectPage(r *http.Request) ProjectPageData {
	project := ProjectByID(r)
	resources := project.GetResources()
	resourceProgress := GetResourceProgress(resources)
	total := float64(resourceProgress.GetTotal())

	if len(resources) > 0 {
		resourceProgress.Done = uint(math.Ceil(float64(resourceProgress.Done) / total * 100))
		resourceProgress.Assigned = uint(math.Ceil(float64(resourceProgress.Assigned) / total * 100))
		resourceProgress.NotEnough = uint(math.Ceil(float64(resourceProgress.NotEnough) / total * 100))
		resourceProgress.Absent = uint(math.Ceil(float64(resourceProgress.Absent) / total * 100))
	} else {
		resourceProgress = ResourceProgress{}
	}

	resourceProgress.Normalize()

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
