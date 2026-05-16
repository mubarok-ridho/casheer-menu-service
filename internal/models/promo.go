package models

import (
	"time"
	"gorm.io/gorm"
)

type Promo struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TenantID    uint           `json:"tenant_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	PromoPrice  float64        `json:"promo_price" gorm:"not null"`
	ImageURL    string         `json:"image_url"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	StartAt     time.Time      `json:"start_at"`
	EndAt       *time.Time     `json:"end_at"`
	StartTime   string         `json:"start_time" gorm:"default:''"`
	EndTime     string         `json:"end_time" gorm:"default:''"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Items []PromoItem `json:"items,omitempty"`
}

// AddonMode mengontrol bagaimana variasi di promo item bekerja:
//   "fixed"   - variasi sudah ditentukan saat setup promo, tidak bisa diubah customer
//   "dynamic" - customer bebas memilih variasi saat checkout, harga = promo_price + delta variasi
//   ""        - menu ini dalam bundle tidak punya variasi
const (
	AddonModeFixed   = "fixed"
	AddonModeDynamic = "dynamic"
)

type PromoItem struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	PromoID     uint   `json:"promo_id" gorm:"not null;index"`
	MenuID      uint   `json:"menu_id" gorm:"not null"`
	VariationID *uint  `json:"variation_id"`
	AddonMode   string `json:"addon_mode" gorm:"default:''"`
	Quantity    int    `json:"quantity" gorm:"default:1"`

	Menu      Menu          `json:"menu,omitempty"`
	Variation MenuVariation `json:"variation,omitempty"`
}
