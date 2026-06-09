package users

import (
	"time"
)

type User struct {
	ID                 string     `json:"id" gorm:"primaryKey;type:uuid"`
	Email              string     `json:"email" gorm:"not null;unique"`
	HashedPassword     string     `json:"-" gorm:"not null"`
	Name               string     `json:"name" gorm:"not null"`
	EmailVerified      bool       `json:"email_verified" gorm:"default:false;not null"`
	VerificationToken  string     `json:"-" gorm:"type:varchar(255)"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
