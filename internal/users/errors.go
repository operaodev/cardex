package users

import "errors"

var (
	// ErrUserNotFound se devuelve cuando no existe un usuario con el email dado.
	ErrUserNotFound = errors.New("usuario no encontrado")

	// ErrEmailAlreadyExists se devuelve cuando el email ya está registrado.
	ErrEmailAlreadyExists = errors.New("el email ya está registrado")

	// ErrInvalidCredentials se devuelve cuando la contraseña no coincide.
	ErrInvalidCredentials = errors.New("Credenciales inválidas")

	// ErrInvalidVerificationToken se devuelve cuando el token de verificación es inválido.
	ErrInvalidVerificationToken = errors.New("Token de verificación inválido")

	// ErrEmailAlreadyVerified se devuelve cuando se intenta verificar un email ya verificado.
	ErrEmailAlreadyVerified = errors.New("El email ya está verificado")

	// ErrPasswordTooShort se devuelve cuando la contraseña es demasiado corta.
	ErrPasswordTooShort = errors.New("La contraseña debe tener al menos 8 caracteres")

	// ErrEmailNotVerified se devuelve cuando el usuario intenta iniciar sesión sin verificar su email.
	ErrEmailNotVerified = errors.New("Debes verificar tu correo electrónico antes de iniciar sesión")
)
