package providers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/operaodev/cardex/internal/products"
)

func TestYGOProvider_FetchItemsByName(t *testing.T) {
	// Set up mock HTTP server
	mux := http.NewServeMux()

	// Mock YGOPRO deck API
	mux.HandleFunc("/cardinfo.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"data": [
				{
					"id": 86982413,
					"name": "Dark Magician",
					"humanReadableCardType": "Normal Monster",
					"desc": "The ultimate wizard in terms of attack and defense.",
					"archetype": "Dark Magician",
					"card_images": [
						{
							"image_url": "https://images.ygoprodeck.com/images/cards/86982413.jpg",
							"image_url_small": "https://images.ygoprodeck.com/images/cards_small/86982413.jpg",
							"image_url_cropped": "https://images.ygoprodeck.com/images/cards_cropped/86982413.jpg"
						}
					]
				}
			]
		}`)
	})

	// Mock Yugipedia card page by numeric ID
	mux.HandleFunc("/wiki/86982413", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html>
		<html>
		<body>
			<!-- Other languages table -->
			<table class="wikitable" style="width: 100%;">
				<tbody>
					<tr>
						<th scope="col">Language</th>
						<th scope="col">Name</th>
						<th scope="col">Card Text</th>
					</tr>
					<tr>
						<th scope="row">English</th>
						<td>Dark Magician</td>
						<td>The ultimate wizard in terms of attack and defense.</td>
					</tr>
					<tr>
						<th scope="row">Spanish</th>
						<td>Mago Oscuro</td>
						<td>El mago definitivo en términos de ataque y defensa.</td>
					</tr>
				</tbody>
			</table>

			<!-- CTS tables -->
			<table class="cts" id="cts--EN">
				<thead>
					<tr>
						<th class="cts__header--number">Number</th>
						<th class="cts__header--set">Release Set</th>
						<th class="cts__header--rarity">Rarity</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>LOB-005</td>
						<td>Legend of Blue Eyes White Dragon</td>
						<td><a href="/wiki/Ultra_Rare">Ultra Rare</a></td>
					</tr>
				</tbody>
			</table>

			<table class="cts" id="cts--SP">
				<thead>
					<tr>
						<th class="cts__header--number">Number</th>
						<th class="cts__header--set-localized">Nombre Localizado</th>
						<th class="cts__header--set">Release Set</th>
						<th class="cts__header--rarity">Rarity</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>LOB-SP005</td>
						<td>Leyenda del Dragón Blanco de Ojos Azules</td>
						<td>Legend of Blue Eyes White Dragon</td>
						<td><a href="/wiki/Ultra_Rare">Ultra Rare</a></td>
					</tr>
				</tbody>
			</table>
		</body>
		</html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Initialize YGOProvider and target the mock server
	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.ygoproBaseUrl = server.URL
	ygoProv.yugipediaBaseUrl = server.URL

	// Call FetchItemsByName
	itemsList, err := ygoProv.FetchItemsByName("Dark Magician")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify itemsList contains expected English and Spanish printed card items
	if len(itemsList) != 2 {
		t.Fatalf("expected 2 items, got %d", len(itemsList))
	}

	// Check fields
	var enItem, spItem products.Product
	for _, item := range itemsList {
		switch item.Lang {
		case products.EN:
			enItem = item
		case products.SP:
			spItem = item
		}
	}

	if enItem.ExternalID != "Dark Magician" || enItem.Code != "LOB-005" || enItem.Name != "Dark Magician" {
		t.Errorf("incorrect English item fields: %+v", enItem)
	}

	if spItem.ExternalID != "Dark Magician" || spItem.Code != "LOB-SP005" || spItem.Name != "Mago Oscuro" {
		t.Errorf("incorrect Spanish item fields: %+v", spItem)
	}

	if spItem.SetName != "Leyenda del Dragón Blanco de Ojos Azules" {
		t.Errorf("expected localized set name: Leyenda del Dragón Blanco de Ojos Azules, got: %s", spItem.SetName)
	}
}
