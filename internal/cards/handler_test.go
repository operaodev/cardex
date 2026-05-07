package cards

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestRouter configura el router de Gin en modo Test y registra las rutas del handler
func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	cardsGroup := r.Group("/cards")
	{
		cardsGroup.GET("/search", handler.GetByNameHandler)
		cardsGroup.GET("/:id", handler.GetByIDHandler)
	}
	
	return r
}

func TestGetByIDHandler(t *testing.T) {
	// 1. Arrange (Preparar datos)
	mockCards := []Card{
		{ID: 12345, Names: map[LangCode]string{"en": "Dark Magician"}},
	}
	mockRepo := NewMockRepository(mockCards)
	svc := NewService(mockRepo)
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	// Crear una petición HTTP falsa
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cards/12345", nil)

	// 2. Act (Ejecutar)
	router.ServeHTTP(w, req)

	// 3. Assert (Verificar resultados)
	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba status 200, se obtuvo %d", w.Code)
	}

	var response Card
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Fallo al parsear el JSON de respuesta: %v", err)
	}

	if response.Names["en"] != "Dark Magician" {
		t.Errorf("Se esperaba nombre 'Dark Magician', se obtuvo '%s'", response.Names["en"])
	}
}

func TestGetByNameHandler(t *testing.T) {
	// 1. Arrange
	mockCards := []Card{
		{ID: 1, Names: map[LangCode]string{"en": "Dark Magician"}},
		{ID: 2, Names: map[LangCode]string{"en": "Dark Magician"}},
	}
	mockRepo := NewMockRepository(mockCards)
	svc := NewService(mockRepo)
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cards/search?name=Dark%20Magician", nil)

	// 2. Act
	router.ServeHTTP(w, req)

	// 3. Assert
	if w.Code != http.StatusOK {
		t.Errorf("Se esperaba status 200, se obtuvo %d", w.Code)
	}

	var response []Card
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Fallo al parsear el JSON de respuesta: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Se esperaban 2 cartas, se obtuvieron %d", len(response))
	}
}
