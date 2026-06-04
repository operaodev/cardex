package providers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/operaodev/cardex/internal/items"
)

func TestParseSetPage_StructureDeck(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/wiki/Structure_Deck:_Fire_Kings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<table class="infobox">
				<tr><th class="infobox-above">Structure Deck: Fire Kings</th></tr>
				<tr><td class="infobox-image"><img src="https://ms.yugipedia.com/SR14-DeckEN.png" /></td></tr>
				<tr><th class="infobox-label">Medium</th><td class="infobox-data"><a href="/wiki/TCG">TCG</a></td></tr>
				<tr><th class="infobox-label">Type</th><td class="infobox-data"><ul><li>Structure Deck</li></ul></td></tr>
				<tr><th class="infobox-label">Number of cards</th><td class="infobox-data">48</td></tr>
				<tr><th class="infobox-label">English</th><td class="infobox-data"><i>Structure Deck: Fire Kings</i></td></tr>
				<tr><th class="infobox-label">French</th><td class="infobox-data"><i>Deck de Structure : Les Rois du Feu</i></td></tr>
				<tr><th class="infobox-label">Spanish</th><td class="infobox-data"><i>Baraja de Estructura: Reyes de Fuego</i></td></tr>
				<tr><th class="infobox-label">Prefix</th><td class="infobox-data">
					<ul><li>SR14-EN (en)</li><li>SR14-FR (fr)</li><li>SR14-SP (sp)</li></ul>
				</td></tr>
			</table>
			<h2><span class="mw-headline" id="Breakdown">Breakdown</span></h2>
			<p>Each <i>Structure Deck: Fire Kings</i> contains:</p>
			<ul>
				<li>1 Preconstructed Deck of 48 cards
					<ul>
						<li>5 <a href="/wiki/Ultra_Rare">Ultra Rares</a></li>
						<li>3 <a href="/wiki/Super_Rare">Super Rares</a></li>
						<li>40 <a href="/wiki/Common">Commons</a></li>
					</ul>
				</li>
				<li>1 Double-sided Playmat</li>
			</ul>
			<ul class="gallery">
				<li class="gallerybox"><a class="image"><img src="https://ms.yugipedia.com/SR14-DeckEN.png" /></a></li>
			</ul>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.yugipediaBaseUrl = server.URL

	// Create test item
	testItem := items.Item{
		Type:           items.ItemTypeCard,
		ExternalID:     "Fire King Avatar Garunix",
		SetExternalID:  "Structure Deck: Fire Kings",
		SetCode:        "SR14-EN004",
		Lang:           items.EN,
		Name:           "Fire King Avatar Garunix",
		QuantityPerSet: 3,
	}

	itemsList, err := ygoProv.FetchSets([]items.Item{testItem})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(itemsList) != 1 {
		t.Fatalf("expected 1 item, got %d", len(itemsList))
	}

	item := itemsList[0]
	t.Logf("Set info: SetType=%q QuantityPerBox=%d SetImage=%q",
		item.SetType, item.QuantityPerBox, item.SetImage)

	if item.SetType != "Structure Deck" {
		t.Errorf("SetType should be 'Structure Deck', got %q", item.SetType)
	}
	if item.QuantityPerBox != 48 {
		t.Errorf("QuantityPerBox should be 48, got %d", item.QuantityPerBox)
	}
}

func TestParseSetCardList_WithQuantity(t *testing.T) {
	mux := http.NewServeMux()

	// Mock main set page
	mux.HandleFunc("/wiki/Structure_Deck:_Fire_Kings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<table class="infobox">
				<tr><th class="infobox-above">Structure Deck: Fire Kings</th></tr>
				<tr><th class="infobox-label">Type</th><td class="infobox-data"><ul><li>Structure Deck</li></ul></td></tr>
				<tr><th class="infobox-label">Number of cards</th><td class="infobox-data">48</td></tr>
				<tr><th class="infobox-label">English</th><td class="infobox-data"><i>Structure Deck: Fire Kings</i></td></tr>
			</table>
			<h2><span class="mw-headline" id="Breakdown">Breakdown</span></h2>
			<p>Each <i>Structure Deck: Fire Kings</i> contains:</p>
			<ul><li>1 Preconstructed Deck of 48 cards</li></ul>
			<div class="set-lists-tabber">
				<div class="tabber">
					<div class="tabbertab" title="English">
						<div class="set-list-tab" data-page="Set Card Lists:Structure Deck: Fire Kings (TCG-EN)"></div>
					</div>
				</div>
			</div>
		</body></html>`)
	})

	// Mock Set Card List page
	mux.HandleFunc("/wiki/Set_Card_Lists:Structure_Deck:_Fire_Kings_(TCG-EN)", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<div class="set-list">
				<table class="wikitable sortable card-list set-list__main">
					<tbody>
						<tr>
							<th scope="col" class="set-list__main__header set-list__main__header--card-number">Card number</th>
							<th scope="col" class="set-list__main__header set-list__main__header--name">Name</th>
							<th scope="col" class="set-list__main__header set-list__main__header--rarity">Rarity</th>
							<th scope="col" class="set-list__main__header set-list__main__header--category">Category</th>
							<th scope="col" class="set-list__main__header set-list__main__header--print">Print</th>
							<th scope="col" class="set-list__main__header set-list__main__header--quantity">Quantity</th>
						</tr>
						<tr>
							<td><a href="/wiki/SR14-EN001">SR14-EN001</a></td>
							<td>"<a href="/wiki/Sacred_Fire_King_Garunix">Sacred Fire King Garunix</a>"</td>
							<td><a href="/wiki/Ultra_Rare">Ultra Rare</a></td>
							<td><a href="/wiki/Effect_Monster">Effect Monster</a></td>
							<td>New</td>
							<td>1</td>
						</tr>
						<tr>
							<td><a href="/wiki/SR14-EN004">SR14-EN004</a></td>
							<td>"<a href="/wiki/Fire_King_Avatar_Garunix">Fire King Avatar Garunix</a>"</td>
							<td><a href="/wiki/Common">Common</a></td>
							<td><a href="/wiki/Effect_Monster">Effect Monster</a></td>
							<td>Reprint</td>
							<td>3</td>
						</tr>
						<tr>
							<td><a href="/wiki/SR14-EN024">SR14-EN024</a></td>
							<td>"<a href="/wiki/Fire_King_Sanctuary">Fire King Sanctuary</a>"</td>
							<td><a href="/wiki/Ultra_Rare">Ultra Rare</a></td>
							<td><a href="/wiki/Continuous_Spell_Card">Continuous Spell Card</a></td>
							<td>New</td>
							<td>1</td>
						</tr>
					</tbody>
				</table>
			</div>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.yugipediaBaseUrl = server.URL

	testItem := items.Item{
		Type:          items.ItemTypeCard,
		ExternalID:    "Fire King Avatar Garunix",
		SetExternalID: "Structure Deck: Fire Kings",
		SetCode:       "SR14-EN004",
		Lang:          items.EN,
	}

	itemsList, err := ygoProv.FetchSets([]items.Item{testItem})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Items enriched: %d", len(itemsList))
	for _, item := range itemsList {
		t.Logf("  Item: Code=%q QuantityPerSet=%d QuantityPerBox=%d SetType=%q",
			item.SetCode, item.QuantityPerSet, item.QuantityPerBox, item.SetType)
	}

	if len(itemsList) != 1 {
		t.Fatalf("expected 1 item, got %d", len(itemsList))
	}

	item := itemsList[0]
	if item.SetType != "Structure Deck" {
		t.Errorf("SetType should be 'Structure Deck', got %q", item.SetType)
	}
	if item.QuantityPerBox != 48 {
		t.Errorf("QuantityPerBox should be 48, got %d", item.QuantityPerBox)
	}
	if item.QuantityPerSet != 3 {
		t.Errorf("QuantityPerSet should be 3, got %d", item.QuantityPerSet)
	}
}

func TestParseSetCardList_WithBonusCards(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/wiki/Set_Card_Lists:Legendary_Modern_Decks_2026_(TCG-EN)", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<h3><span class="mw-headline" id="Bonus_cards">Bonus cards</span></h3>
			<div class="set-list">
				<table class="wikitable sortable card-list set-list__main">
					<tbody>
						<tr>
							<th scope="col" class="set-list__main__header set-list__main__header--card-number">Card number</th>
							<th scope="col" class="set-list__main__header set-list__main__header--name">Name</th>
							<th scope="col" class="set-list__main__header set-list__main__header--rarity">Rarity</th>
							<th scope="col" class="set-list__main__header set-list__main__header--category">Category</th>
							<th scope="col" class="set-list__main__header set-list__main__header--print">Print</th>
						</tr>
						<tr>
							<td><a href="/wiki/L26D-ENS01">L26D-ENS01</a></td>
							<td>"<a href="/iki/Sky_Striker_Ace_-_Raye">Sky Striker Ace - Raye</a>"</td>
							<td><a href="/wiki/Secret_Rare">Secret Rare</a></td>
							<td><a href="/wiki/Effect_Monster">Effect Monster</a></td>
							<td>Reprint</td>
						</tr>
					</tbody>
				</table>
			</div>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.yugipediaBaseUrl = server.URL

	testItem := items.Item{
		Type:          items.ItemTypeCard,
		ExternalID:    "Sky Striker Ace - Raye",
		SetExternalID: "Legendary Modern Decks 2026",
		SetCode:       "L26D-ENS01",
		Lang:          items.EN,
	}

	itemsList, err := ygoProv.FetchSets([]items.Item{testItem})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Items enriched: %d", len(itemsList))
	for _, item := range itemsList {
		t.Logf("  Item: Code=%q QuantityPerSet=%d", item.SetCode, item.QuantityPerSet)
	}
}
