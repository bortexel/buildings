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
	Priority      int         `json:"priority"`
	ScreenshotURL null.String `json:"screenshot_url"`
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
	Done      float64
	Assigned  float64
	NotEnough float64
	Absent    float64
}

func GetResourceProgress(resources []Resource) ResourceProgress {
	resourceProgress := ResourceProgress{}

	for _, resource := range resources {
		switch resource.Status {
		case StatusDone:
			resourceProgress.Done += float64(resource.Amount)
		case StatusAssigned:
			resourceProgress.Assigned += float64(resource.Amount)
		case StatusNotEnough:
			resourceProgress.NotEnough += float64(resource.Amount)
		case StatusAbsent:
			resourceProgress.Absent += float64(resource.Amount)
		}
	}

	return resourceProgress
}

func (r *ResourceProgress) NormalizePercentage() {
	for r.GetTotal() > 100 {
		if r.Absent > r.Assigned {
			r.Absent--
			continue
		}

		if r.Assigned > r.NotEnough {
			r.Assigned--
			continue
		}

		if r.NotEnough > r.Done {
			r.NotEnough--
			continue
		}

		if r.Done > 1 {
			r.Done--
			continue
		}
	}
}

func (r *ResourceProgress) GetTotal() float64 {
	return r.Done + r.Assigned + r.NotEnough + r.Absent
}

func ProjectPage(r *http.Request) ProjectPageData {
	project := ProjectByID(r)
	resources := project.GetResources()
	resourceProgress := GetResourceProgress(resources)
	total := resourceProgress.GetTotal()

	if len(resources) > 0 {
		resourceProgress.Done = math.Ceil(resourceProgress.Done / total * 100)
		resourceProgress.Assigned = math.Ceil(resourceProgress.Assigned / total * 100)
		resourceProgress.NotEnough = math.Ceil(resourceProgress.NotEnough / total * 100)
		resourceProgress.Absent = math.Ceil(resourceProgress.Absent / total * 100)
	} else {
		resourceProgress = ResourceProgress{}
	}

	resourceProgress.NormalizePercentage()

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
	Database.Order("priority").Find(&projects)
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
