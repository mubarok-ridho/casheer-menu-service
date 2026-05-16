package models

import (
	"time"
	"gorm.io/gorm"
)

type Ingredient struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TenantID    uint           `json:"tenant_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"not null"`
	Unit        string         `json:"unit" gorm:"not null"`
	Stock       float64        `json:"stock" gorm:"default:0"`
	CostPerUnit float64        `json:"cost_per_unit" gorm:"default:0"`
	LowStockAt  float64        `json:"low_stock_at" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type MenuIngredient struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	MenuID       uint      `json:"menu_id" gorm:"not null;index"`
	IngredientID uint      `json:"ingredient_id" gorm:"not null"`
	Amount       float64   `json:"amount" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Ingredient Ingredient `json:"ingredient,omitempty"`
}
