package models

import "time"

// Category is a user-defined label/tag for todos (e.g. Work, Personal,
// Learning). Categories are scoped per user and unique by name within a user.
type Category struct {
	ID        uint      `json:"id"        gorm:"primaryKey"`
	UserID    uint      `json:"user_id"   gorm:"not null;uniqueIndex:idx_user_category_name"`
	Name      string    `json:"name"      gorm:"size:100;not null;uniqueIndex:idx_user_category_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Todos []Todo `json:"-" gorm:"many2many:todo_categories"`
}

// TableName pins the table name.
func (Category) TableName() string { return "categories" }
