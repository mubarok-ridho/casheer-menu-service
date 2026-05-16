package models

import "time"

// VariationIngredient — bahan TAMBAHAN untuk variasi tertentu (di atas base menu)
type VariationIngredient struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	VariationID  uint      `json:"variation_id" gorm:"not null;index"`
	IngredientID uint      `json:"ingredient_id" gorm:"not null"`
	Amount       float64   `json:"amount" gorm:"not null"` // dalam unit dasar
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Ingredient Ingredient `json:"ingredient,omitempty"`
}
