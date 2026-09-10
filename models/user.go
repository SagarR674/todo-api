package models

import "time"

// User is a registered account. The password field holds a bcrypt hash and is
// never serialised to JSON (`json:"-"`).
type User struct {
	ID        uint      `json:"id"        gorm:"primaryKey"`
	Name      string    `json:"name"      gorm:"size:255;not null"`
	Email     string    `json:"email"     gorm:"size:255;not null;uniqueIndex"`
	Password  string    `json:"-"         gorm:"size:255;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Todos      []Todo     `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	Categories []Category `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

// TableName pins the table name regardless of GORM's pluralisation settings.
func (User) TableName() string { return "users" }
