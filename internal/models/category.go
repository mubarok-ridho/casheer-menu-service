package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TenantID    uint           `json:"tenant_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Menus []Menu `json:"menus,omitempty"`
}
