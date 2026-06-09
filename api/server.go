package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/api/handler"
	"github.com/operaodev/cardex/api/middleware"
	"github.com/operaodev/cardex/internal/stock"
)

type Server struct {
	router             *gin.Engine
	providersHandler   *handler.ProviderHandler
	cardsHandler       *handler.CardsHandler
	usersHandler       *handler.UsersHandler
	syncHandler        *handler.SyncHandler
	productsHandler    *handler.ProductsHandler
	stockHandler       *handler.StockHandler
	marketplaceHandler *handler.MarketplaceHandler
	wishlistHandler    *handler.WishlistHandler
	stockRepo          stock.Repository
	jwtSecret          string
}

func NewServer(
	providersH *handler.ProviderHandler,
	cardsH *handler.CardsHandler,
	usersH *handler.UsersHandler,
	syncH *handler.SyncHandler,
	productsH *handler.ProductsHandler,
	stockH *handler.StockHandler,
	marketplaceH *handler.MarketplaceHandler,
	wishlistH *handler.WishlistHandler,
	stockRepo stock.Repository,
	jwtSecret string,
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
		router:             router,
		providersHandler:   providersH,
		cardsHandler:       cardsH,
		usersHandler:       usersH,
		syncHandler:        syncH,
		productsHandler:    productsH,
		stockHandler:       stockH,
		marketplaceHandler: marketplaceH,
		wishlistHandler:    wishlistH,
		stockRepo:          stockRepo,
		jwtSecret:          jwtSecret,
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
		// GET /users/verify?token=...
		usersGroup.GET("/verify", s.usersHandler.VerifyEmail)
		// GET /users/me (requiere auth)
		usersGroup.GET("/me", middleware.AuthMiddleware(s.jwtSecret), s.usersHandler.GetMe)
	}

	// Rutas de products
	productsGroup := s.router.Group("/products")
	{
		// Las rutas fijas deben ir antes del comodín :id
		// POST /products/suggestions
		productsGroup.POST("/suggestions", s.productsHandler.FindSuggestions)
		// GET /products/random/:count
		productsGroup.GET("/random/:count", s.productsHandler.GetRandomNames)
		// POST /products/related
		productsGroup.POST("/related", s.productsHandler.GetRelatedCards)
		// GET /products/:id
		productsGroup.GET("/:id", s.productsHandler.GetByID)
	}

	stockGroup := s.router.Group("/stock")
	auth := middleware.AuthMiddleware(s.jwtSecret)
	ownership := middleware.RequireStockOwnership(s.stockRepo)
	{
		// GET /stock/me — stock del usuario autenticado
		stockGroup.GET("/me", auth, s.stockHandler.GetMyStock)
		// GET /stock/:user_id
		stockGroup.GET("/:user_id", auth, s.stockHandler.GetByUserID)
		// GET /stock/id/:id — requiere ownership
		stockGroup.GET("/id/:id", auth, ownership, s.stockHandler.GetByID)
		// GET /stock/logs/:stock_id — requiere ownership
		stockGroup.GET("/logs/:stock_id", auth, ownership, s.stockHandler.GetLogs)
		// POST /stock
		stockGroup.POST("", auth, s.stockHandler.Create)
		// POST /stock/restock
		stockGroup.POST("/restock", auth, s.stockHandler.Restock)
		// POST /stock/return
		stockGroup.POST("/return", auth, s.stockHandler.Return)
		// POST /stock/sale
		stockGroup.POST("/sale", auth, s.stockHandler.Sale)
		// POST /stock/trade
		stockGroup.POST("/trade", auth, s.stockHandler.Trade)
		// POST /stock/gift
		stockGroup.POST("/gift", auth, s.stockHandler.Gift)
		// POST /stock/lost
		stockGroup.POST("/lost", auth, s.stockHandler.Lost)
		// POST /stock/damage
		stockGroup.POST("/damage", auth, s.stockHandler.Damage)
		// POST /stock/adjust
		stockGroup.POST("/adjust", auth, s.stockHandler.Adjust)
		// POST /stock/rollback
		stockGroup.POST("/rollback", auth, s.stockHandler.Rollback)
	}

	// Marketplace analysis
	marketplaceGroup := s.router.Group("/marketplace")
	{
		// GET /marketplace/analysis/:id
		marketplaceGroup.GET("/analysis/:id", s.marketplaceHandler.GetPrices)
		// GET /marketplace/offers/:id
		marketplaceGroup.GET("/offers/:id", s.marketplaceHandler.GetOffers)
	}

	// Wishlist (custom packs)
	wishlistGroup := s.router.Group("/wishlist")
	{
		// GET /wishlist
		wishlistGroup.GET("", auth, s.wishlistHandler.GetMyWishlist)
		// POST /wishlist
		wishlistGroup.POST("", auth, s.wishlistHandler.Upsert)
		// DELETE /wishlist/:product_id
		wishlistGroup.DELETE("/:product_id", auth, s.wishlistHandler.Delete)
	}
}

func (s *Server) Start(addr string) error {
	log.Printf("Iniciando servidor en %s", addr)
	return s.router.Run(addr)
}
