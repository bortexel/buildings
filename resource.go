package main

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"strconv"
	"strings"
)

const (
	StatusAbsent    = 0
	StatusNotEnough = 1
	StatusAssigned  = 2
	StatusDone      = 3
)

type Resource struct {
	PrimaryKey
	Name        string      `json:"name" gorm:"not null"`
	MinecraftID null.String `json:"minecraft_id"`
	Amount      uint        `json:"amount" gorm:"not null"`
	Status      uint        `json:"status" gorm:"not null"`
	ProjectID   uint        `json:"project_id" gorm:"not null"`
	AssigneeID  null.Int    `json:"assignee_id"`
	Timestamps
}

func ListResources(_ *http.Request) (interface{}, error) {
	var resources []Resource
	Database.Find(&resources)
	return resources, nil
}

func (r Resource) StatusText() string {
	switch r.Status {
	case StatusAbsent:
		return "Отсутствует"
	case StatusNotEnough:
		return "Частично есть"
	case StatusAssigned:
		return "В процессе"
	case StatusDone:
		return "Готово"
	default:
		return ""
	}
}

func (r Resource) TableClass() string {
	switch r.Status {
	case StatusAbsent:
		return "table-danger"
	case StatusNotEnough:
		return "table-warning"
	case StatusAssigned:
		return "table-info"
	case StatusDone:
		return "table-success"
	default:
		return ""
	}
}

func (r Resource) AmountText() string {
	stacks := r.Amount / 64
	remaining := r.Amount - stacks*64

	builder := strings.Builder{}
	if stacks > 0 {
		builder.WriteString(fmt.Sprintf("%d ст.", stacks))

		if remaining > 0 {
			builder.WriteString(" + ")
		}
	}

	if remaining > 0 {
		builder.WriteString(strconv.Itoa(int(remaining)))
	}

	return builder.String()
}

func (r Resource) GetAssigneeName() string {
	if r.AssigneeID.Valid {
		var member Member
		Database.Find(&member, r.AssigneeID.Int64)
		if member.IsValid() {
			return member.Name
		}
	}

	return ""
}
