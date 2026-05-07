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
			ID:           1,
			Names:        map[cards.LangCode]string{"sp": "Mago Oscuro"},
			Descriptions: map[cards.LangCode]string{"sp": "El mago supremo."},
		},
		{
			ID:           2,
			Names:        map[cards.LangCode]string{"en": "Ojama Black"},
			Descriptions: map[cards.LangCode]string{"en": "It is very weak, but it can be used as a shield."},
		},
		{
			ID:           3,
			Names:        map[cards.LangCode]string{"sp": "Ojama Negro"},
			Descriptions: map[cards.LangCode]string{"sp": "Es muy débil, pero se puede usar como escudo."},
		},
		{
			ID:           4,
			Names:        map[cards.LangCode]string{"en": "Dark Magician"},
			Descriptions: map[cards.LangCode]string{"en": "El mago supremo."},
		},
		{
			ID:           5,
			Names:        map[cards.LangCode]string{"en": "Blue-Eyes White Dragon"},
			Descriptions: map[cards.LangCode]string{"en": "This legendary dragon is a powerful engine of destruction."},
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
		// /cards/search/:provider/:id   (por id) /cards/search/ygo/1234
		cardsGroup.GET("/search/:provider/:id", searchHandler.SearchByIDInProvider)
		// /cards/search/:provider?name=Kuriboh
		cardsGroup.GET("/search/:provider", searchHandler.SearchByNamesInProvider)
	}

	log.Println("Servidor iniciado en http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
