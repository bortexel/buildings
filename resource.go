package main

type Resource struct {
	PrimaryKey
	MinecraftID string `json:"minecraft_id"`
	Amount      uint   `json:"amount"`
	Status      uint   `json:"status"`
	ProjectID   uint   `json:"project_id"`
	Timestamps
}
