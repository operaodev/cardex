package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/jwt"
	"github.com/operaodev/cardex/internal/users"
)

// AuthResponse es la respuesta pública tras registro o login.
type AuthResponse struct {
	User  *users.User `json:"user"`
	Token string      `json:"token"`
}

// UsersHandler expone las funciones del servicio de usuarios a través de HTTP.
type UsersHandler struct {
	service     users.Service
	jwtSecret   string
	jwtDuration time.Duration
}

// NewUsersHandler crea una nueva instancia del Handler inyectando el servicio.
func NewUsersHandler(s users.Service, jwtSecret string, jwtDuration time.Duration) *UsersHandler {
	return &UsersHandler{service: s, jwtSecret: jwtSecret, jwtDuration: jwtDuration}
}

// Register maneja las peticiones de registro de nuevos usuarios.
// Body JSON: { "name": "...", "email": "...", "password": "..." }
func (h *UsersHandler) Register(c *gin.Context) {
	var input users.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidJSONBody})
		return
	}

	user, err := h.service.Register(input)
	if err != nil {
		if errors.Is(err, users.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, user.Name, h.jwtSecret, h.jwtDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar el token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{User: user, Token: token})
}

// Login maneja las peticiones de autenticación de usuarios.
// Body JSON: { "email": "...", "password": "..." }
func (h *UsersHandler) Login(c *gin.Context) {
	var input users.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidJSONBody})
		return
	}

	user, err := h.service.Login(input)
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
		if errors.Is(err, users.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error interno del servidor"})
		return
	}

	token, err := jwt.GenerateToken(user.ID, user.Email, user.Name, h.jwtSecret, h.jwtDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar el token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{User: user, Token: token})
}

// VerifyEmail maneja la verificación del correo electrónico mediante el token.
// Query params: token
func (h *UsersHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token de verificación requerido"})
		return
	}

	user, err := h.service.VerifyEmail(token)
	if err != nil {
		if errors.Is(err, users.ErrInvalidVerificationToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, users.ErrEmailAlreadyVerified) {
			c.JSON(http.StatusOK, gin.H{"message": "El correo ya está verificado", "user": user})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar el correo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Correo verificado exitosamente", "user": user})
}

// GetMe devuelve el perfil del usuario autenticado.
// GET /users/me
func (h *UsersHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("userID")
	user, err := h.service.GetByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al obtener el perfil"})
		return
	}
	c.JSON(http.StatusOK, user)
}
