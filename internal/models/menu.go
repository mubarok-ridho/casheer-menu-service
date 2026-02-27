package models

import (
    "time"
    "gorm.io/gorm"
)

type Menu struct {
    ID           uint           `gorm:"primarykey" json:"id"`
    TenantID     uint           `json:"tenant_id"`
    CategoryID   uint           `json:"category_id"`
    Name         string         `json:"name"`
    Description  string         `json:"description"`
    BasePrice    float64        `json:"base_price"`
    IsAvailable  bool           `json:"is_available" gorm:"default:true"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
    Category     Category       `json:"category,omitempty"`
    Variations   []MenuVariation `json:"variations,omitempty"`
    Images       []MenuImage    `json:"images,omitempty"`
    OrderCount   int64          `json:"order_count" gorm:"-:all"` // Untuk best seller
}

type MenuVariation struct {
    ID        uint    `gorm:"primarykey" json:"id"`
    MenuID    uint    `json:"menu_id"`
    Name      string  `json:"name"` // Contoh: "Ukuran", "Level Pedas"
    Option    string  `json:"option"` // Contoh: "Sedang", "Extra Hot"
    Price     float64 `json:"price"` // Additional price
    Stock     int     `json:"stock" gorm:"default:0"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type MenuImage struct {
    ID        uint   `gorm:"primarykey" json:"id"`
    MenuID    uint   `json:"menu_id"`
    ImageURL  string `json:"image_url"`
    PublicID  string `json:"public_id"` // Untuk Cloudinary
    IsPrimary bool   `json:"is_primary" gorm:"default:false"`
    CreatedAt time.Time `json:"created_at"`
}