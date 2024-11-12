package models

import (
	"time"

	_ "github.com/lib/pq"
)

type Task struct {
	ID          int        `gorm:"primaryKey;autoIncrement;column:id"`
	Title       *string    `gorm:"type:varchar;column:title"`
	Description *string    `gorm:"type:varchar;column:description"`
	Status      *string    `gorm:"type:varchar;column:status;default:'pending'"`
	DueDate     *time.Time `gorm:"column:due_date"`
	CreatedAt   *time.Time `gorm:"column:created_at"`
	CreatedBy   *string    `gorm:"type:varchar;column:created_by"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
	UpdatedBy   *string    `gorm:"type:varchar;column:updated_by"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

type User struct {
	ID       int     `gorm:"primaryKey;autoIncrement;column:id"`
	Username *string `gorm:"type:varchar;column:username"`
	Password *string `gorm:"type:varchar;column:password"` // to do hashed
}
