package main

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Model interface {
	IsValid() bool
}

type PrimaryKey struct {
	ID uint `gorm:"primaryKey" json:"id"`
}

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt null.Time `gorm:"index" json:"deleted_at"`
}

func (p PrimaryKey) IsValid() bool {
	return p.ID > 0
}
