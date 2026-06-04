package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/api/handler"
)

type Server struct {
	router           *gin.Engine
	providersHandler *handler.ProviderHandler
	cardsHandler     *handler.CardsHandler
	usersHandler     *handler.UsersHandler
	inventoryHandler *handler.InventoryHandler
	syncHandler      *handler.SyncHandler
	itemsHandler     *handler.ItemsHandler
}

func NewServer(
	providersH *handler.ProviderHandler,
	cardsH *handler.CardsHandler,
	usersH *handler.UsersHandler,
	inventoryH *handler.InventoryHandler,
	syncH *handler.SyncHandler,
	itemsH *handler.ItemsHandler,
) *Server {
	router := gin.Default()

	// Middleware de CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	s := &Server{
		router:           router,
		providersHandler: providersH,
		cardsHandler:     cardsH,
		usersHandler:     usersH,
		inventoryHandler: inventoryH,
		syncHandler:      syncH,
		itemsHandler:     itemsH,
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
	}

	providersGroup := s.router.Group("/providers")
	{
		// GET /providers/:provider/cards
		providersGroup.GET("/:provider/cards", s.providersHandler.FetchCards)
		// GET /providers/:provider/cards/:name
		providersGroup.GET("/:provider/cards/:name", s.providersHandler.FetchCardsByName)
	}

	// Rutas de sincronización (administración)
	syncGroup := s.router.Group("/sync")
	{
		// GET /sync/status
		syncGroup.GET("/status", s.syncHandler.SyncStatus)
		// POST /sync/ygo
		syncGroup.POST("/:tcg", s.syncHandler.TriggerSync)
		// POST /sync/ygo/by-name
		syncGroup.POST("/:tcg/by-name", s.syncHandler.TriggerSyncByName)
	}

	// Rutas de usuarios (auth)
	usersGroup := s.router.Group("/users")
	{
		// POST /users/register
		usersGroup.POST("/register", s.usersHandler.Register)
		// POST /users/login
		usersGroup.POST("/login", s.usersHandler.Login)
	}

	// Rutas de ítems
	itemsGroup := s.router.Group("/items")
	{
		// Las rutas fijas deben ir antes del comodín :id
		// POST /items/suggestions
		itemsGroup.POST("/suggestions", s.itemsHandler.FindSuggestions)
		// GET /items/random/:count
		itemsGroup.GET("/random/:count", s.itemsHandler.GetRandomNames)
		// GET /items/:id
		itemsGroup.GET("/:id", s.itemsHandler.GetByID)
	}

	invGroup := s.router.Group("/inventory")
	{
		// GET /inventory/:user_id
		invGroup.GET("/:user_id", s.inventoryHandler.GetInventory)
		// GET /inventory/logs/:inventory_id
		invGroup.GET("/logs/:inventory_id", s.inventoryHandler.GetLogs)
		// POST /inventory/restock
		invGroup.POST("/restock", s.inventoryHandler.Restock)
		// POST /inventory/sell
		invGroup.POST("/sell", s.inventoryHandler.Sell)
		// POST /inventory/loss
		invGroup.POST("/loss", s.inventoryHandler.RegisterLoss)
		// POST /inventory/return
		invGroup.POST("/return", s.inventoryHandler.RegisterReturn)
		// POST /inventory/price
		invGroup.POST("/price", s.inventoryHandler.ChangePrice)
	}
}

func (s *Server) Start(addr string) error {
	log.Printf("Iniciando servidor en %s", addr)
	return s.router.Run(addr)
}
