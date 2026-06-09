package users

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository define los métodos que nuestra capa de datos de usuarios debe tener.
type Repository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	FindByVerificationToken(token string) (*User, error)
	UpdateName(id, name string) error
	UpdatePassword(id, hashedPassword string) error
	VerifyEmail(id string, verifiedAt time.Time) error
	SetVerificationToken(id, token string) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository crea una nueva instancia del repositorio de usuarios.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create ejecuta un INSERT INTO users con los datos del usuario ya validados.
// El campo hashed_password debe venir pre-hasheado desde la capa de servicio.
func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByEmail busca un usuario por su dirección de correo electrónico.
// A nivel SQL ejecuta un SELECT * FROM users WHERE email = ? LIMIT 1.
// Devuelve un error específico si el usuario no existe.
func (r *repository) FindByEmail(email string) (*User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindByID busca un usuario por su ID.
func (r *repository) FindByID(id string) (*User, error) {
	var user User
	result := r.db.Where("id = ?", id).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindByVerificationToken busca un usuario por su token de verificación de email.
func (r *repository) FindByVerificationToken(token string) (*User, error) {
	var user User
	result := r.db.Where("verification_token = ?", token).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidVerificationToken
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// UpdateName actualiza el nombre del usuario.
func (r *repository) UpdateName(id, name string) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Update("name", name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdatePassword actualiza la contraseña hasheada del usuario.
func (r *repository) UpdatePassword(id, hashedPassword string) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Update("hashed_password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// VerifyEmail marca el email del usuario como verificado.
func (r *repository) VerifyEmail(id string, verifiedAt time.Time) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"email_verified":      true,
		"verified_at":         verifiedAt,
		"verification_token":  "",
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetVerificationToken asigna un token de verificación al usuario.
func (r *repository) SetVerificationToken(id, token string) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Update("verification_token", token)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
