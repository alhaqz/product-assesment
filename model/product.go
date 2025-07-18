package model

import "time"

type Product struct {
	ProductID   int64      `gorm:"primaryKey;autoIncrement"`
	Name        string     `gorm:"type:varchar(500)"`
	Price       float64    `json:"price"`
	Description string     `gorm:"type:varchar(500)"`
	Quantity    int32      `json:"quantity"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
}
