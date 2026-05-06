package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/operaodev/cardex/internal/cards"
	"github.com/operaodev/cardex/internal/search"
	searchproviders "github.com/operaodev/cardex/internal/search/providers"
)

func main() {
	// database.Connect()
	
	mockCards := []cards.Card{
		{
			ID:          "YGO-123-SP",
			Lang:        "SP",
			Name:        "Mago Oscuro",
			Description: "El mago supremo.",
		},
		{
			ID:          "YGO-125-EN",
			Lang:        "EN",
			Name:        "Ojama Black",
			Description: "It is very weak, but it can be used as a shield.",
		},
		{
			ID:          "YGO-125-SP",
			Lang:        "SP",
			Name:        "Ojama Negro",
			Description: "Es muy débil, pero se puede usar como escudo.",
		},
		{
			ID:          "YGO-123-EN",
			Lang:        "EN",
			Name:        "Dark Magician",
			Description: "El mago supremo.",
		},
		{
			ID:          "YGO-001-EN",
			Lang:        "EN",
			Name:        "Blue-Eyes White Dragon",
			Description: "This legendary dragon is a powerful engine of destruction.",
		},
	}

	repo := cards.NewMockRepository(mockCards)
	service := cards.NewService(repo)
	handler := cards.NewHandler(service)

	ygoProv := searchproviders.NewYGOProvider()
	searchSvc := search.NewService(ygoProv)
	searchHandler := search.NewHandler(searchSvc)

	r := gin.Default()

	cardsGroup := r.Group("/cards")
	{
		// /cards/search?name=Kuriboh
		cardsGroup.GET("/search", handler.GetByNameHandler)
		// /cards/scg-1234
		cardsGroup.GET("/:id", handler.GetByIDHandler)
		// /cards/search/provider/ygo/id/1234
		cardsGroup.GET("/provider/:provider/:id", searchHandler.GetFromProvider)
	}

	log.Println("Servidor iniciado en http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
