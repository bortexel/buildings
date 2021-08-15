package main

import "gopkg.in/guregu/null.v4"

const (
	StatusNotActive = 0
	StatusBusy      = 1
	StatusActive    = 2
)

type Member struct {
	PrimaryKey
	Name   string      `json:"name" gorm:"not null"`
	Roles  int         `json:"roles" gorm:"not null"`
	Status int         `json:"status" gorm:"not null"`
	Note   null.String `json:"note"`
	Timestamps
}

func (m Member) GetNote() string {
	if m.Note.Valid {
		return m.Note.String
	}
	return ""
}

func (m Member) StatusText() string {
	switch m.Status {
	case StatusNotActive:
		return "Неактивен"
	case StatusBusy:
		return "Нет возможности"
	case StatusActive:
		return "Активен"
	default:
		return ""
	}
}

func (m Member) TableClass() string {
	switch m.Status {
	case StatusNotActive:
		return "table-danger"
	case StatusBusy:
		return "table-warning"
	case StatusActive:
		return "table-success"
	default:
		return ""
	}
}

func (m Member) IsDesigner() bool {
	return m.Roles&0b001 > 0
}

func (m Member) IsBuilder() bool {
	return m.Roles&0b010 > 0
}

func (m Member) IsProvider() bool {
	return m.Roles&0b100 > 0
}

func AllMembers() []Member {
	var members []Member
	Database.Find(&members)
	return members
}
