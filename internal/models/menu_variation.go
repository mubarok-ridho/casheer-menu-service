package models

import (
	"time"

	"gorm.io/gorm"
)

type MenuVariation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	MenuID    uint           `json:"menu_id" gorm:"not null;index"`
	Name      string         `json:"name" gorm:"not null"`   // Contoh: "Ukuran", "Level Pedas"
	Option    string         `json:"option" gorm:"not null"` // Contoh: "Sedang", "Extra Hot"
	Price     float64        `json:"price" gorm:"default:0"` // Additional price
	Stock     int            `json:"stock" gorm:"default:0"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Menu Menu `json:"menu,omitempty"`
}
