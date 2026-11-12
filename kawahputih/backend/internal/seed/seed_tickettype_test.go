package seed

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/KawahPutih/backend/internal/domain/tickettype"
)

// openTestDB meniru setup SQLite yang dipakai tes domain lain.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&tickettype.TicketType{}); err != nil {
		t.Fatalf("migrate ticket_types: %v", err)
	}
	return db
}

func TestSeedTicketTypesMatchesPriceList(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := seedTicketTypes(ctx, db); err != nil {
		t.Fatalf("seedTicketTypes: %v", err)
	}

	want := map[string]struct {
		name  string
		price int64
	}{
		"HTM-WNI-REG":           {"HTM Wisatawan Nusantara", 31000},
		"HTM-WNA-REG":           {"HTM Wisatawan Mancanegara", 90000},
		"HTM-WNI-SUNRISE":       {"HTM Wisatawan Nusantara (Sunrise)", 46000},
		"HTM-WNA-SUNRISE":       {"HTM Wisatawan Mancanegara (Sunrise)", 105000},
		"LHKP-REG":              {"Charge LHKP (Reguler)", 179000},
		"LHKP-SUNRISE":          {"Charge LHKP (Sunrise)", 179000},
		"ONTANG-ANTING":         {"Ontang Anting", 32000},
		"ONTANG-ANTING-SUNRISE": {"Ontang Anting (Sunrise)", 38000},
		"VEHICLE-RODA2":         {"Kendaraan Roda 2", 10000},
		"VEHICLE-RODA4":         {"Kendaraan Roda 4", 18000},
		"VEHICLE-RODA6":         {"Kendaraan Roda 6", 48000},
		"HELM-STORAGE":          {"Jasa Penitipan Helm", 3000},
		"JASA-FOTOGRAFER":       {"Jasa Fotografer", 11000},
		"KUDA-TUNGGANG":         {"Kuda Tunggang", 55000},
		"PANAHAN":               {"Panahan", 22000},
		"SUNAN-IBU":             {"Sunan Ibu", 15000},
		"JEMBATAN-APUNG":        {"Jembatan Apung", 17000},
		"SKYWALK-CANOPY":        {"Skywalk Canopy", 17000},
		"PREWEDDING":            {"Prewedding", 550000},
		"SHOOTING":              {"Shooting", 3300000},
	}

	var got []tickettype.TicketType
	if err := db.WithContext(ctx).Find(&got).Error; err != nil {
		t.Fatalf("load ticket types: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("seeded %d ticket types, want %d", len(got), len(want))
	}
	seen := map[string]bool{}
	for _, tt := range got {
		exp, ok := want[tt.Code]
		if !ok {
			t.Errorf("unexpected ticket code %q", tt.Code)
			continue
		}
		seen[tt.Code] = true
		if tt.Name != exp.name {
			t.Errorf("%s: name %q, want %q", tt.Code, tt.Name, exp.name)
		}
		if tt.Price != exp.price {
			t.Errorf("%s: price %d, want %d", tt.Code, tt.Price, exp.price)
		}
		if !tt.IsActive {
			t.Errorf("%s: expected active", tt.Code)
		}
		// Setiap kategori yang di-seed harus salah satu yang diterima API.
		valid := false
		for _, c := range tickettype.Categories {
			if c == tt.Category {
				valid = true
				break
			}
		}
		if !valid {
			t.Errorf("%s: category %q not in tickettype.Categories", tt.Code, tt.Category)
		}
	}
	for code := range want {
		if !seen[code] {
			t.Errorf("missing seeded ticket %q", code)
		}
	}
}
func TestSeedTicketTypesRefreshesExistingRows(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := seedTicketTypes(ctx, db); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	// Simulasikan baris usang seperti dari katalog lama.
	if err := db.WithContext(ctx).Model(&tickettype.TicketType{}).
		Where("code = ?", "HTM-WNI-REG").
		Updates(map[string]any{
			"name":      "Tiket Masuk Umum",
			"price":     int64(15000),
			"category":  "ENTRANCE",
			"is_active": false,
		}).Error; err != nil {
		t.Fatalf("mutate row: %v", err)
	}

	if err := seedTicketTypes(ctx, db); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	var refreshed tickettype.TicketType
	if err := db.WithContext(ctx).Where("code = ?", "HTM-WNI-REG").First(&refreshed).Error; err != nil {
		t.Fatalf("reload row: %v", err)
	}
	if refreshed.Name != "HTM Wisatawan Nusantara" {
		t.Errorf("name not refreshed: %q", refreshed.Name)
	}
	if refreshed.Price != 31000 {
		t.Errorf("price not refreshed: %d, want 31000", refreshed.Price)
	}
	if !refreshed.IsActive {
		t.Error("row not re-activated by seed")
	}

	// Menjalankan ulang tidak boleh menggandakan baris.
	var count int64
	if err := db.WithContext(ctx).Model(&tickettype.TicketType{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 20 {
		t.Errorf("row count %d after reseed, want 20", count)
	}
}
