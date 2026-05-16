package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	TenantID      uint           `json:"tenant_id" gorm:"not null;index"`
	OrderNumber   string         `json:"order_number" gorm:"uniqueIndex;not null"`
	CustomerName  string         `json:"customer_name"`
	CustomerPhone string         `json:"customer_phone"`
	TotalAmount   float64        `json:"total_amount" gorm:"not null"`
	PaymentMethod string         `json:"payment_method"`
	PaymentStatus string         `json:"payment_status" gorm:"default:'pending'"`
	OrderStatus   string         `json:"order_status" gorm:"default:'pending'"`
	Notes          string         `json:"notes"`
	DiscountType   string         `json:"discount_type"`
	DiscountAmount float64        `json:"discount_amount"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Items []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	OrderID     uint      `json:"order_id" gorm:"not null;index"`
	MenuID      uint      `json:"menu_id" gorm:"not null"`
	VariationID *uint     `json:"variation_id"`
	PromoID     *uint     `json:"promo_id"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	Price       float64   `json:"price" gorm:"not null"`
	Subtotal    float64   `json:"subtotal" gorm:"not null"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Order     Order         `json:"order,omitempty"`
	Menu      Menu          `json:"menu,omitempty"`
	Variation MenuVariation `json:"variation,omitempty"`
}
