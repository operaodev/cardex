package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/api/handler"
)

type Server struct {
	router        *gin.Engine
	cardsHandler  *handler.CardsHandler
	searchHandler *handler.SearchHandler
	syncHandler   *handler.SyncHandler
}

func NewServer(cardsH *handler.CardsHandler, searchH *handler.SearchHandler, syncH *handler.SyncHandler) *Server {
	s := &Server{
		router:        gin.Default(),
		cardsHandler:  cardsH,
		searchHandler: searchH,
		syncHandler:   syncH,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	cardsGroup := s.router.Group("/cards")
	{
		// GET /cards?tcg=ygo&lang=sp&page=1&limit=20
		cardsGroup.GET("", s.cardsHandler.GetCatalog)
		// GET /cards/1234
		cardsGroup.GET("/:id", s.cardsHandler.GetByID)
		// GET /cards/suggestions?tcg=ygo&lang=sp&name=Kuriboh
		cardsGroup.GET("/suggestions", s.cardsHandler.GetSuggestions)

		// Proveedores externos
		// GET /cards/search/:provider/:id
		cardsGroup.GET("/search/:provider/:id", s.searchHandler.SearchByIDInProvider)
		// GET /cards/search/:provider?name=Kuriboh
		cardsGroup.GET("/search/:provider", s.searchHandler.SearchByNamesInProvider)
		// GET /cards/search/:provider/all
		cardsGroup.GET("/search/:provider/all", s.searchHandler.SearchAllInProvider)
	}

	// Rutas de sincronización (administración)
	syncGroup := s.router.Group("/sync")
	{
		// GET /sync/status
		syncGroup.GET("/status", s.syncHandler.SyncStatus)
		// POST /sync/ygo
		syncGroup.POST("/:tcg", s.syncHandler.TriggerSync)
	}
}

func (s *Server) Start(addr string) error {
	log.Printf("Iniciando servidor en %s", addr)
	return s.router.Run(addr)
}
