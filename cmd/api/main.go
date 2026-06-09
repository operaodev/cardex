package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/operaodev/cardex/api"
	"github.com/operaodev/cardex/api/handler"
	"github.com/operaodev/cardex/internal/cards"
	custompacks "github.com/operaodev/cardex/internal/custom_packs"
	"github.com/operaodev/cardex/internal/database"
	"github.com/operaodev/cardex/internal/mailer"
	"github.com/operaodev/cardex/internal/marketplace"
	"github.com/operaodev/cardex/internal/products"
	"github.com/operaodev/cardex/internal/providers"
	"github.com/operaodev/cardex/internal/stock"
	syncsvc "github.com/operaodev/cardex/internal/sync"
	"github.com/operaodev/cardex/internal/users"
)

func main() {
	// 0. Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: no se cargó .env, usando variables del sistema")
	}

	// 1. Inicializar Base de Datos (Conexión y Automigración)
	database.Connect()

	// 2. Inicializar Repositorios (Capa de Datos)
	repo := cards.NewRepository(database.DB)
	productsRepo := products.NewRepository(database.DB)

	// 3. Inicializar Servicios (Lógica de Negocio)
	cardsSvc := cards.NewService(repo)
	productsSvc := products.NewService(productsRepo)

	ygoProv := providers.NewYGOProvider()
	providerSvc := providers.NewService(ygoProv)

	syncService := syncsvc.NewSyncService(providerSvc, productsRepo)

	// 4. Inicializar Handlers (Capa de Transporte)
	cardsHandler := handler.NewCardsHandler(cardsSvc)
	productsHandler := handler.NewProductsHandler(productsSvc)
	providerHandler := handler.NewProviderHandler(providerSvc)
	syncHandler := handler.NewSyncHandler(syncService)

	// Inicializar Mailer SMTP
	smtpMailer := mailer.NewSMTPMailer(
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASS"),
		os.Getenv("SMTP_FROM"),
		os.Getenv("APP_URL"),
	)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET es obligatorio")
	}
	jwtDuration := 24 * time.Hour
	if d, err := time.ParseDuration(os.Getenv("JWT_DURATION")); err == nil {
		jwtDuration = d
	}

	usersRepo := users.NewRepository(database.DB)
	usersSvc := users.NewService(usersRepo, smtpMailer)
	usersHandler := handler.NewUsersHandler(usersSvc, jwtSecret, jwtDuration)

	stockRepo := stock.NewRepository(database.DB)
	stockSvc := stock.NewService(stockRepo)
	stockHandler := handler.NewStockHandler(stockSvc)

	marketplaceRepo := marketplace.NewRepository(database.DB)
	marketplaceSvc := marketplace.NewService(marketplaceRepo)
	marketplaceHandler := handler.NewMarketplaceHandler(marketplaceSvc)

	wishlistRepo := custompacks.NewRepository(database.DB)
	wishlistSvc := custompacks.NewService(wishlistRepo)
	wishlistHandler := handler.NewWishlistHandler(wishlistSvc)

	// 5. Configurar e Iniciar Servidor
	srv := api.NewServer(
		providerHandler,
		cardsHandler,
		usersHandler,
		syncHandler,
		productsHandler,
		stockHandler,
		marketplaceHandler,
		wishlistHandler,
		stockRepo,
		jwtSecret,
	)

	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
