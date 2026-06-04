package providers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/operaodev/cardex/internal/items"
)

// TestParseCTSTables_RealStructure tests with the actual HTML structure
// from Yugipedia (with release column, set-localized, etc.)
func TestParseCTSTables_RealStructure(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/cardinfo.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"data": [
				{
					"id": 86982413,
					"name": "Dark Magician",
					"humanReadableCardType": "Normal Monster",
					"desc": "The ultimate wizard.",
					"archetype": "Dark Magician",
					"card_images": [{"image_url": "https://img.com/dm.jpg", "image_url_small": "", "image_url_cropped": ""}]
				}
			]
		}`)
	})

	// Mock Yugipedia with REAL structure (release column, no <thead>, cts__header classes)
	mux.HandleFunc("/wiki/86982413", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html>
		<html>
		<body>
			<table class="wikitable" style="width: 100%;">
				<tbody>
					<tr>
						<th scope="col">Language</th>
						<th scope="col">Name</th>
						<th scope="col">Card Text</th>
					</tr>
					<tr><th scope="row">English</th><td>Dark Magician</td><td>The ultimate wizard.</td></tr>
					<tr><th scope="row">Spanish</th><td>Mago Oscuro</td><td>El mago definitivo.</td></tr>
					<tr><th scope="row">Japanese</th><td>ブラック・マジシャン</td><td>攻撃力・守備力最強の魔法使い。</td></tr>
				</tbody>
			</table>

			<!-- CTS EN with release column -->
			<table id="cts--EN" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2004-03-01</td><td>SYE-001</td><td>Starter Deck: Yugi Evolution</td><td><a href="/wiki/SR">Super Rare</a></td></tr>
					<tr><td>2004-10-12</td><td>DB1-EN102</td><td>Dark Beginning 1</td><td><a href="/wiki/UR">Ultra Rare</a></td></tr>
				</tbody>
			</table>

			<!-- CTS NA (regional, should map to EN) -->
			<table id="cts--NA" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2002-03-08</td><td>LOB-005</td><td>Legend of Blue Eyes White Dragon</td><td><a href="/wiki/UR">Ultra Rare</a></td></tr>
				</tbody>
			</table>

			<!-- CTS SP with set-localized -->
			<table id="cts--SP" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set-localized">Nombre</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2005-01-01</td><td>LOB-SP005</td><td>Leyenda BEWD</td><td>Legend of Blue Eyes</td><td><a href="/wiki/UR">Ultra Rare</a></td></tr>
				</tbody>
			</table>

			<!-- CTS JP with EMPTY code (old print) -->
			<table id="cts--JP" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--set-localized">Japanese name</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>1999-05-27</td><td></td><td>Vol.1</td><td>Ｖｏｌ．１</td><td><a href="/wiki/UR">Ultra Rare</a></td></tr>
					<tr><td>2000-04-20</td><td>DM1-001</td><td>DM1</td><td>デュエルモンスターズ</td><td><a href="/wiki/C">Common</a></td></tr>
				</tbody>
			</table>

			<!-- CTS KR (no translation in wikitable — should fallback to EN) -->
			<table id="cts--KR" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2006-01-01</td><td>LOB-KR005</td><td>Legend of Blue Eyes</td><td><a href="/wiki/UR">Ultra Rare</a></td></tr>
				</tbody>
			</table>
		</body>
		</html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.ygoproBaseUrl = server.URL
	ygoProv.yugipediaBaseUrl = server.URL

	itemsList, err := ygoProv.FetchCardsByName("Dark Magician")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Total items produced: %d", len(itemsList))
	for _, item := range itemsList {
		t.Logf("  Lang=%s Code=%q Set=%q Rarity=%q Name=%q", item.Lang, item.Code, item.SetName, item.Rarity, item.Name)
	}

	// Expected: 2 EN + 1 NA(→EN) + 1 SP + 2 JP + 1 KR(fallback→EN) = 7 items
	if len(itemsList) != 7 {
		t.Errorf("expected 7 items, got %d", len(itemsList))
	}

	// Verify JP Vol.1 with empty code is captured (Bug #3 fix)
	var jpEmptyCode bool
	for _, item := range itemsList {
		if item.Lang == items.JP && item.Code == "" {
			jpEmptyCode = true
		}
	}
	if !jpEmptyCode {
		t.Error("JP print with empty code (Vol.1) was not captured — Bug #3 not fixed")
	}

	// Verify KR uses EN fallback name (Bug #1 fix)
	var krItem items.Item
	for _, item := range itemsList {
		if item.Lang == items.KR {
			krItem = item
		}
	}
	if krItem.Name != "Dark Magician" {
		t.Errorf("KR print should fallback to EN name 'Dark Magician', got %q", krItem.Name)
	}

	// Verify NA regional table was captured (Bug #2 fix)
	var naFound bool
	for _, item := range itemsList {
		if item.Code == "LOB-005" && item.Lang == items.EN {
			naFound = true
		}
	}
	if !naFound {
		t.Error("NA regional print (LOB-005) was not captured — Bug #2 not fixed")
	}
}

// TestScrapeCards_Disambiguation tests the retry logic when a card name
// leads to an archetype/set page instead of the card page.
func TestScrapeCards_Disambiguation(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/cardinfo.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"data": [{
				"id": 12345678,
				"name": "Purrely",
				"humanReadableCardType": "Effect Monster",
				"desc": "Test card.",
				"card_images": [{"image_url": "https://img.com/p.jpg", "image_url_small": "", "image_url_cropped": ""}]
			}]
		}`)
	})

	// /wiki/12345678 is the actual card page with CTS tables
	mux.HandleFunc("/wiki/12345678", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<table class="wikitable" style="width: 100%;">
				<tbody>
					<tr>
						<th scope="col">Language</th>
						<th scope="col">Name</th>
						<th scope="col">Card Text</th>
					</tr>
					<tr><th scope="row">English</th><td>Purrely</td><td>Test card.</td></tr>
				</tbody>
			</table>
			<table id="cts--EN" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2023-02-10</td><td>AMDE-EN013</td><td>Amazing Defenders</td><td><a href="/wiki/SR">Super Rare</a></td></tr>
				</tbody>
			</table>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.ygoproBaseUrl = server.URL
	ygoProv.yugipediaBaseUrl = server.URL

	itemsList, err := ygoProv.FetchCardsByName("Purrely")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(itemsList) != 1 {
		t.Fatalf("expected 1 item (from disambiguation retry), got %d", len(itemsList))
	}

	if itemsList[0].Code != "AMDE-EN013" {
		t.Errorf("expected code AMDE-EN013, got %q", itemsList[0].Code)
	}
	t.Logf("Disambiguation retry worked: %s %s %s", itemsList[0].Name, itemsList[0].Code, itemsList[0].Rarity)
}

func TestScrapeCards_GalleryImages(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/cardinfo.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"data": [{
				"id": 62357126,
				"name": "Danger!? Jackalope?",
				"humanReadableCardType": "Effect Monster",
				"desc": "Test card.",
				"archetype": "Danger!",
				"card_images": [{"image_url": "https://img.com/jack.jpg", "image_url_small": "", "image_url_cropped": ""}]
			}]
		}`)
	})

	mux.HandleFunc("/wiki/62357126", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<div class="hlist" style="text-align: center;">
				<ul>
					<li><a href="/wiki/Card_Gallery:Danger!%3F_Jackalope%3F" title="Card Gallery:Danger!? Jackalope?">Gallery</a></li>
					<li><a href="/wiki/Card_Tips:Danger!%3F_Jackalope%3F" title="Card Tips:Danger!? Jackalope?">Tips</a></li>
				</ul>
			</div>
			<table class="wikitable" style="width: 100%;">
				<tbody>
					<tr><th scope="col">Language</th><th scope="col">Name</th><th scope="col">Card Text</th></tr>
					<tr><th scope="row">Spanish</th><td>¿¡Peligro!? ¿Lebrílope?</td><td>Texto en español.</td></tr>
				</tbody>
			</table>
			<table id="cts--EN" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2019-08-29</td><td><a href="/wiki/MP19-EN139">MP19-EN139</a></td><td><a href="/wiki/2019_Gold_Sarcophagus_Tin_Mega_Pack"><i>2019 Gold Sarcophagus Tin Mega Pack</i></a></td><td><a href="/wiki/Prismatic_Secret_Rare">Prismatic Secret Rare</a></td></tr>
				</tbody>
			</table>
			<table id="cts--SP" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--set-localized">Spanish name</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2019-08-29</td><td><a href="/wiki/MP19-SP139">MP19-SP139</a></td><td><a href="/wiki/2019_Gold_Sarcophagus_Tin_Mega_Pack"><i>2019 Gold Sarcophagus Tin Mega Pack</i></a></td><td>Lata Cofre de Oro Sellado 2019 Mega Pack</td><td><a href="/wiki/Prismatic_Secret_Rare">Prismatic Secret Rare</a></td></tr>
				</tbody>
			</table>
		</body></html>`)
	})

	mux.HandleFunc("/wiki/Card_Gallery:Danger!%3F_Jackalope%3F", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<div id="card-gallery--EN" class="card-gallery">
				<h2><span class="mw-headline">Worldwide English</span></h2>
				<ul class="gallery mw-gallery-traditional">
					<li class="gallerybox" style="width: 155px">
						<div style="width: 155px">
							<div class="thumb" style="width: 150px;">
								<div style="margin:15px auto;">
									<a href="/wiki/File:DangerJackalope-MP19-EN-PScR-1E.png" class="image">
										<img alt="" src="https://ms.yugipedia.com//thumb/f/f6/DangerJackalope-MP19-EN-PScR-1E.png/120px-DangerJackalope-MP19-EN-PScR-1E.png" width="120" height="175" />
									</a>
								</div>
							</div>
							<div class="gallerytext">
								<p><a href="/wiki/MP19-EN139">MP19-EN139</a> (<a href="/wiki/Prismatic_Secret_Rare" title="Prismatic Secret Rare">PScR</a>)<br /><a href="/wiki/1st_Edition">1st Edition</a><br /><i><a href="/wiki/2019_Gold_Sarcophagus_Tin_Mega_Pack">2019 Gold Sarcophagus Tin Mega Pack</a></i></p>
							</div>
						</div>
					</li>
					<li class="gallerybox" style="width: 155px">
						<div style="width: 155px">
							<div class="thumb" style="width: 150px;">
								<div style="margin:15px auto;">
									<a href="/wiki/File:DangerJackalope-MP19-EN-PScR-1E-OP.png" class="image">
										<img alt="" src="https://ms.yugipedia.com//thumb/proxy.png/120px-proxy.png" width="120" height="175" />
									</a>
								</div>
							</div>
							<div class="gallerytext">
								<p><a href="/wiki/MP19-EN139">MP19-EN139</a> (<a href="/wiki/Prismatic_Secret_Rare" title="Prismatic Secret Rare">PScR</a>)<br /><a href="/wiki/1st_Edition">1st Edition</a><br /><a href="/wiki/Official_Proxy">Official Proxy</a><br /><i><a href="/wiki/2019_Gold_Sarcophagus_Tin_Mega_Pack">2019 Gold Sarcophagus Tin Mega Pack</a></i></p>
							</div>
						</div>
					</li>
				</ul>
			</div>
			<div id="card-gallery--SP" class="card-gallery">
				<h2><span class="mw-headline">Spanish</span></h2>
				<ul class="gallery mw-gallery-traditional">
					<li class="gallerybox" style="width: 155px">
						<div style="width: 155px">
							<div class="thumb" style="width: 150px;">
								<div style="margin:15px auto;">
									<a href="/wiki/File:DangerJackalope-MP19-SP-PScR-1E.png" class="image">
										<img alt="" src="https://ms.yugipedia.com//thumb/a/ab/DangerJackalope-MP19-SP-PScR-1E.png/120px-DangerJackalope-MP19-SP-PScR-1E.png" width="120" height="175" />
									</a>
								</div>
							</div>
							<div class="gallerytext">
								<p><a href="/wiki/MP19-SP139">MP19-SP139</a> (<a href="/wiki/Prismatic_Secret_Rare" title="Prismatic Secret Rare">PScR</a>)<br /><a href="/wiki/1st_Edition">1st Edition</a><br /><i><a href="/wiki/2019_Gold_Sarcophagus_Tin_Mega_Pack">Lata Cofre de Oro 2019</a></i></p>
							</div>
						</div>
					</li>
					<li class="gallerybox" style="width: 155px">
						<div style="width: 155px">
							<div class="thumb" style="height: 205px;">DangerJackalope-MP19-SP-BROKEN.png</div>
							<div class="gallerytext">
								<p><a href="/wiki/MP19-SP139">MP19-SP139</a> (<a href="/wiki/Prismatic_Secret_Rare" title="Prismatic Secret Rare">PScR</a>)<br /><a href="/wiki/1st_Edition">1st Edition</a></p>
							</div>
						</div>
					</li>
				</ul>
			</div>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.ygoproBaseUrl = server.URL
	ygoProv.yugipediaBaseUrl = server.URL

	itemsList, err := ygoProv.FetchCardsByName("Danger!? Jackalope?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Total items produced: %d", len(itemsList))
	for _, item := range itemsList {
		t.Logf("  Lang=%s Code=%q Rarity=%q PrintURLSmall=%q PrintURLLarge=%q Edition=%q RarityCode=%q",
			item.Lang, item.Code, item.Rarity, item.PrintURLSmall, item.PrintURLLarge, item.Edition, item.RarityCode)
	}

	// Expected: 1 EN + 1 SP = 2 items
	if len(itemsList) != 2 {
		t.Errorf("expected 2 items, got %d", len(itemsList))
	}

	// Find EN item
	var enItem items.Item
	for _, item := range itemsList {
		if item.Lang == items.EN {
			enItem = item
			break
		}
	}

	// Verify EN item has gallery data
	if enItem.PrintURLSmall == "" {
		t.Error("EN item should have PrintURLSmall from gallery")
	}
	if enItem.PrintURLLarge == "" {
		t.Error("EN item should have PrintURLLarge from gallery")
	}
	if enItem.Edition != "1st Edition" {
		t.Errorf("EN item Edition should be '1st Edition', got %q", enItem.Edition)
	}
	if enItem.RarityCode != "PScR" {
		t.Errorf("EN item RarityCode should be 'PScR', got %q", enItem.RarityCode)
	}

	// Verify PrintURLLarge is derived correctly
	expectedLarge := "https://ms.yugipedia.com//f/f6/DangerJackalope-MP19-EN-PScR-1E.png"
	if enItem.PrintURLLarge != expectedLarge {
		t.Errorf("EN PrintURLLarge should be %q, got %q", expectedLarge, enItem.PrintURLLarge)
	}

	// Find SP item
	var spItem items.Item
	for _, item := range itemsList {
		if item.Lang == items.SP {
			spItem = item
			break
		}
	}

	// Verify SP item has gallery data
	if spItem.PrintURLSmall == "" {
		t.Error("SP item should have PrintURLSmall from gallery")
	}
	if spItem.PrintURLLarge == "" {
		t.Error("SP item should have PrintURLLarge from gallery")
	}
	if spItem.Edition != "1st Edition" {
		t.Errorf("SP item Edition should be '1st Edition', got %q", spItem.Edition)
	}
	if spItem.RarityCode != "PScR" {
		t.Errorf("SP item RarityCode should be 'PScR', got %q", spItem.RarityCode)
	}

	t.Logf("Gallery scraping worked: EN=%s, SP=%s", enItem.Code, spItem.Code)
}

func TestScrapeCards_DangerJackalope(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/cardinfo.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"data": [{
				"id": 62357126,
				"name": "Danger!? Jackalope?",
				"humanReadableCardType": "Effect Monster",
				"desc": "You can reveal this card in your hand...",
				"archetype": "Danger!",
				"card_images": [{"image_url": "https://img.com/jack.jpg", "image_url_small": "", "image_url_cropped": ""}]
			}]
		}`)
	})

	mux.HandleFunc("/wiki/62357126", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body>
			<div class="hlist" style="text-align: center;">
				<ul>
					<li><a href="/wiki/Card_Gallery:Danger!%3F_Jackalope%3F" title="Card Gallery:Danger!? Jackalope?">Gallery</a></li>
					<li><a href="/wiki/Card_Tips:Danger!%3F_Jackalope%3F" title="Card Tips:Danger!? Jackalope?">Tips</a></li>
				</ul>
			</div>
			<table class="wikitable" style="width: 100%;">
				<tbody>
					<tr>
						<th scope="col">Language</th>
						<th scope="col">Name</th>
						<th scope="col">Card Text</th>
					</tr>
					<tr><th scope="row">French</th><td><span lang="fr">Danger !? Jackalope ?</span></td><td><span lang="fr">Texte en français.</span></td></tr>
					<tr><th scope="row">German</th><td><span lang="de">Gefahr!? Wolpertinger?</span></td><td><span lang="de">Deutscher Text.</span></td></tr>
					<tr><th scope="row">Italian</th><td><span lang="it">Pericolo!? Jackalope?</span></td><td><span lang="it">Testo in italiano.</span></td></tr>
					<tr><th scope="row">Portuguese</th><td><span lang="pt">Perigo!? Lebrílope?</span></td><td><span lang="pt">Texto em português.</span></td></tr>
					<tr><th scope="row">Spanish</th><td><span lang="es">¿¡Peligro!? ¿Lebrílope?</span></td><td><span lang="es">Texto en español.</span></td></tr>
					<tr><th scope="row">Japanese</th><td><span lang="ja">未界域のジャッカロープ</span></td><td><span lang="ja">日本語テキスト。</span></td></tr>
					<tr><th scope="row">Korean</th><td><span lang="ko">미계역의 재카로프</span></td><td><span lang="ko">한국어 텍스트.</span></td></tr>
					<tr><th scope="row">Simplified Chinese</th><td><span lang="zh-Hans">未界域之鹿角兔</span></td><td><span lang="zh-Hans">简体中文文本。</span></td></tr>
				</tbody>
			</table>

			<!-- TCG EN: 1 print with 1 rarity + 1 print with 3 rarities = 4 items -->
			<table id="cts--EN" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2018-07-26</td><td><a href="/wiki/CYHO-EN085">CYHO-EN085</a></td><td><a href="/wiki/Cybernetic_Horizon"><i>Cybernetic Horizon</i></a></td><td><a href="/wiki/Ultra_Rare">Ultra Rare</a></td></tr>
					<tr><td>2023-11-02</td><td><a href="/wiki/RA01-EN013">RA01-EN013</a></td><td><a href="/wiki/25th_Anniversary_Rarity_Collection"><i>25th Anniversary Rarity Collection</i></a></td><td><a href="/wiki/Super_Rare">Super Rare</a><br /><a href="/wiki/Ultra_Rare">Ultra Rare</a><br /><a href="/wiki/Secret_Rare">Secret Rare</a></td></tr>
				</tbody>
			</table>

			<!-- TCG SP: 1 print with localized set name = 1 item -->
			<table id="cts--SP" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--set-localized">Spanish name</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2018-07-26</td><td><a href="/wiki/CYHO-SP085">CYHO-SP085</a></td><td><a href="/wiki/Cybernetic_Horizon"><i>Cybernetic Horizon</i></a></td><td>Horizonte Cibernético</td><td><a href="/wiki/Ultra_Rare">Ultra Rare</a></td></tr>
				</tbody>
			</table>

			<!-- OCG JP: 1 print with localized set name = 1 item -->
			<table id="cts--JP" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--set-localized">Japanese name</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2019-09-14</td><td><a href="/wiki/EP19-JP024">EP19-JP024</a></td><td><a href="/wiki/Extra_Pack_2019"><i>Extra Pack 2019</i></a></td><td>ＥＸＴＲＡ ＰＡＣＫ ２０１９</td><td><a href="/wiki/Ultra_Rare">Ultra Rare</a></td></tr>
				</tbody>
			</table>

			<!-- OCG SC: 1 print with 3 rarities = 3 items -->
			<table id="cts--SC" class="wikitable sortable card-list cts">
				<tbody>
					<tr>
						<th scope="col" class="cts__header--release">Release</th>
						<th scope="col" class="cts__header--number">Number</th>
						<th scope="col" class="cts__header--set">Set</th>
						<th scope="col" class="cts__header--set-localized">Simplified Chinese name</th>
						<th scope="col" class="cts__header--rarity">Rarity</th>
					</tr>
					<tr><td>2021-08-21</td><td><a href="/wiki/MGP2-SC107">MGP2-SC107</a></td><td><a href="/wiki/Mega_Pack_02"><i>Mega Pack 02</i></a></td><td>超级包02</td><td><a href="/wiki/Ultra_Rare">Ultra Rare</a><br /><a href="/wiki/Secret_Rare">Secret Rare</a><br /><a href="/wiki/Prismatic_Secret_Rare">Prismatic Secret Rare</a></td></tr>
				</tbody>
			</table>

			<!-- Non-language wikitable (limitation history) — should be IGNORED by parseCards -->
			<table class="wikitable">
				<tbody>
					<tr><th scope="col">Status</th><th scope="col">Start date</th><th scope="col">End date</th></tr>
					<tr><td>Unlimited</td><td>2018-07-26</td><td></td></tr>
				</tbody>
			</table>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ygoProv := NewYGOProvider()
	ygoProv.httpClient = server.Client()
	ygoProv.ygoproBaseUrl = server.URL
	ygoProv.yugipediaBaseUrl = server.URL

	itemsList, err := ygoProv.FetchCardsByName("Danger!? Jackalope?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Total items produced: %d", len(itemsList))
	for _, item := range itemsList {
		t.Logf("  Lang=%s Code=%q Set=%q Rarity=%q Name=%q", item.Lang, item.Code, item.SetName, item.Rarity, item.Name)
	}

	// EN: CYHO-EN085 (1 rarity) + RA01-EN013 (3 rarities) = 4
	// SP: CYHO-SP085 (1 rarity) = 1
	// JP: EP19-JP024 (1 rarity) = 1
	// SC: MGP2-SC107 (3 rarities) = 3
	// Total: 9
	if len(itemsList) != 9 {
		t.Errorf("expected 9 items, got %d", len(itemsList))
	}

	var spItem items.Item
	for _, item := range itemsList {
		if item.Lang == items.SP {
			spItem = item
			break
		}
	}
	if spItem.Name != "¿¡Peligro!? ¿Lebrílope?" {
		t.Errorf("SP name should be '¿¡Peligro!? ¿Lebrílope?', got %q", spItem.Name)
	}
	if spItem.SetName != "Horizonte Cibernético" {
		t.Errorf("SP set name should be 'Horizonte Cibernético', got %q", spItem.SetName)
	}

	var jpItem items.Item
	for _, item := range itemsList {
		if item.Lang == items.JP {
			jpItem = item
			break
		}
	}
	if jpItem.Name != "未界域のジャッカロープ" {
		t.Errorf("JP name should be '未界域のジャッカロープ', got %q", jpItem.Name)
	}

	var scCount int
	for _, item := range itemsList {
		if item.Lang == items.SC {
			scCount++
		}
	}
	if scCount != 3 {
		t.Errorf("expected 3 SC items (3 rarities), got %d", scCount)
	}

	var enRARarity bool
	for _, item := range itemsList {
		if item.Lang == items.EN && item.Code == "RA01-EN013" && item.Rarity == "Secret Rare" {
			enRARarity = true
		}
	}
	if !enRARarity {
		t.Error("EN RA01-EN013 should have 'Secret Rare' from multi-rarity print")
	}
}
