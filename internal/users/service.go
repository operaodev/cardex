package users

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/operaodev/cardex/internal/mailer"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInput contiene los datos necesarios para registrar un nuevo usuario.
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput contiene las credenciales para iniciar sesión.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangePasswordInput contiene los datos para cambiar la contraseña.
type ChangePasswordInput struct {
	UserID      string `json:"-"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangeNameInput contiene los datos para cambiar el nombre.
type ChangeNameInput struct {
	UserID string `json:"-"`
	Name   string `json:"name"`
}

// Service define el contrato de lo que nuestra aplicación puede hacer con los usuarios.
type Service interface {
	Register(input RegisterInput) (*User, error)
	Login(input LoginInput) (*User, error)
	GetByID(id string) (*User, error)
	GenerateVerificationToken(userID string) (string, error)
	VerifyEmail(token string) (*User, error)
	ChangePassword(input ChangePasswordInput) error
	ChangeName(input ChangeNameInput) (*User, error)
}

// service implementa la interfaz Service e inyecta el repositorio.
type service struct {
	repo   Repository
	mailer mailer.Mailer
}

// NewService crea una nueva instancia del servicio de usuarios.
func NewService(repo Repository, mailer mailer.Mailer) Service {
	return &service{repo: repo, mailer: mailer}
}

// Register valida los datos, hashea la contraseña y crea el usuario en la base de datos.
// Genera un token de verificación y envía un email al usuario.
// Devuelve error si el email ya está registrado o si los datos son inválidos.
func (s *service) Register(input RegisterInput) (*User, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	if input.Email == "" {
		return nil, fmt.Errorf("el email es obligatorio")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("el nombre es obligatorio")
	}
	if len(input.Password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// Verificar si el email ya está en uso
	existing, err := s.repo.FindByEmail(input.Email)
	if err != nil && err != ErrUserNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hashear la contraseña con bcrypt (cost=12 es un buen balance de seguridad/rendimiento)
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("error al procesar la contraseña: %w", err)
	}

	// Generar token de verificación
	token := generateRandomToken(32)

	user := &User{
		ID:                uuid.NewString(),
		Name:              input.Name,
		Email:             input.Email,
		HashedPassword:    string(hashed),
		VerificationToken: token,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("error al crear el usuario: %w", err)
	}

	// Enviar email de verificación
	if err := s.mailer.SendVerificationEmail(user.Email, user.Name, token); err != nil {
		return nil, fmt.Errorf("error al enviar email de verificación: %w", err)
	}

	return user, nil
}

// Login verifica las credenciales del usuario y devuelve el usuario si son correctas.
// Solo permite login si el email está verificado.
// Nunca revela si el email no existe (devuelve ErrInvalidCredentials en ambos casos)
// para evitar enumeración de usuarios.
func (s *service) Login(input LoginInput) (*User, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		// Enmascarar ErrUserNotFound para no revelar si el email existe
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	return user, nil
}

// GetByID obtiene un usuario por su ID.
func (s *service) GetByID(id string) (*User, error) {
	return s.repo.FindByID(id)
}

// GenerateVerificationToken genera un token aleatorio para verificar el email del usuario.
func (s *service) GenerateVerificationToken(userID string) (string, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return "", err
	}

	if user.EmailVerified {
		return "", ErrEmailAlreadyVerified
	}

	token := generateRandomToken(32)
	if err := s.repo.SetVerificationToken(userID, token); err != nil {
		return "", err
	}

	return token, nil
}

// VerifyEmail verifica el email del usuario usando el token proporcionado.
func (s *service) VerifyEmail(token string) (*User, error) {
	if token == "" {
		return nil, ErrInvalidVerificationToken
	}

	user, err := s.repo.FindByVerificationToken(token)
	if err != nil {
		return nil, err
	}

	if user.EmailVerified {
		return user, ErrEmailAlreadyVerified
	}

	verifiedAt := time.Now().UTC()
	if err := s.repo.VerifyEmail(user.ID, verifiedAt); err != nil {
		return nil, err
	}

	user.EmailVerified = true
	user.VerifiedAt = &verifiedAt
	user.VerificationToken = ""

	return user, nil
}

// ChangePassword cambia la contraseña del usuario después de validar la contraseña actual.
func (s *service) ChangePassword(input ChangePasswordInput) error {
	if len(input.NewPassword) < 8 {
		return ErrPasswordTooShort
	}

	user, err := s.repo.FindByID(input.UserID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.OldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("error al procesar la contraseña: %w", err)
	}

	return s.repo.UpdatePassword(user.ID, string(hashed))
}

// ChangeName actualiza el nombre del usuario.
func (s *service) ChangeName(input ChangeNameInput) (*User, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, fmt.Errorf("el nombre es obligatorio")
	}

	if err := s.repo.UpdateName(input.UserID, input.Name); err != nil {
		return nil, err
	}

	return s.repo.FindByID(input.UserID)
}

// generateRandomToken genera un token aleatorio hexadecimal de la longitud especificada.
func generateRandomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback a UUID si hay error con crypto/rand
		return uuid.NewString()
	}
	return hex.EncodeToString(bytes)
}
