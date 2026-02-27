package models

import (
	"time"

	"gorm.io/gorm"
)

type Menu struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TenantID    uint           `json:"tenant_id" gorm:"not null;index"`
	CategoryID  uint           `json:"category_id" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	BasePrice   float64        `json:"base_price" gorm:"not null"`
	IsAvailable bool           `json:"is_available" gorm:"default:true"`
	PrepTime    int            `json:"prep_time"` // in minutes
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Category   Category        `json:"category,omitempty"`
	Variations []MenuVariation `json:"variations,omitempty"`
	Images     []MenuImage     `json:"images,omitempty"`

	// Calculated fields (not in DB)
	OrderCount int64 `json:"order_count" gorm:"-"`
}
