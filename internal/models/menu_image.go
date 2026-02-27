package models

import (
	"time"

	"gorm.io/gorm"
)

type MenuImage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	MenuID    uint           `json:"menu_id" gorm:"not null;index"`
	ImageURL  string         `json:"image_url" gorm:"not null"`
	PublicID  string         `json:"public_id"` // For Cloudinary
	IsPrimary bool           `json:"is_primary" gorm:"default:false"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Menu Menu `json:"menu,omitempty"`
}
