package main

import (
	"net/http"
)

type Resource struct {
	PrimaryKey
	Name        string `json:"name"`
	MinecraftID string `json:"minecraft_id"`
	Amount      uint   `json:"amount"`
	Status      uint   `json:"status"`
	ProjectID   uint   `json:"project_id"`
	Timestamps
}

func ListResources(_ *http.Request) (interface{}, error) {
	var resources []Resource
	Database.Find(&resources)
	return resources, nil
}
