package seed

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/KawahPutih/backend/internal/domain/article"
	"github.com/KawahPutih/backend/internal/domain/destinationinfo"
	"github.com/KawahPutih/backend/internal/domain/facility"
	"github.com/KawahPutih/backend/internal/domain/faq"
	"github.com/KawahPutih/backend/internal/domain/gallery"
	tourismpackage "github.com/KawahPutih/backend/internal/domain/package"
	"github.com/KawahPutih/backend/internal/domain/payments"
	"github.com/KawahPutih/backend/internal/domain/role"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/internal/domain/visitorrule"
)

func Run(ctx context.Context, db *gorm.DB, log *zap.Logger) error {
	if err := seedRoles(ctx, db); err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}
	if err := seedUsers(ctx, db); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedTicketTypes(ctx, db); err != nil {
		return fmt.Errorf("seed ticket types: %w", err)
	}
	if err := seedPackages(ctx, db); err != nil {
		return fmt.Errorf("seed packages: %w", err)
	}
	if err := seedArticles(ctx, db); err != nil {
		return fmt.Errorf("seed articles: %w", err)
	}
	if err := seedGallery(ctx, db); err != nil {
		return fmt.Errorf("seed gallery: %w", err)
	}
	if err := seedDestinationInfo(ctx, db); err != nil {
		return fmt.Errorf("seed destination info: %w", err)
	}
	if err := seedFAQs(ctx, db); err != nil {
		return fmt.Errorf("seed faqs: %w", err)
	}
	if err := seedFacilities(ctx, db); err != nil {
		return fmt.Errorf("seed facilities: %w", err)
	}
	if err := seedVisitorRules(ctx, db); err != nil {
		return fmt.Errorf("seed visitor rules: %w", err)
	}
	if err := seedPaymentDestinations(ctx, db); err != nil {
		return fmt.Errorf("seed payment destinations: %w", err)
	}
	log.Info("seed data ready")
	return nil
}

// seedPaymentDestinations mengisi daftar awal tujuan pembayaran (tanpa nomor
// rekening) agar admin tinggal mengisi nomor dan nama pemiliknya.
func seedPaymentDestinations(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&payments.PaymentDestination{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaults := []payments.PaymentDestination{
		{MethodCode: payments.MethodQRIS, ProviderName: "QRIS (semua e-wallet & mobile banking)", Active: false, SortOrder: 1},
		{MethodCode: payments.MethodBCAVA, ProviderName: "BCA", Active: true, SortOrder: 10},
		{MethodCode: payments.MethodBRIVA, ProviderName: "BRI", Active: true, SortOrder: 11},
		{MethodCode: payments.MethodMandiriVA, ProviderName: "Mandiri", Active: true, SortOrder: 12},
		{MethodCode: payments.MethodBNIVA, ProviderName: "BNI", Active: true, SortOrder: 13},
		{MethodCode: payments.MethodGoPay, ProviderName: "GoPay", Active: true, SortOrder: 20},
		{MethodCode: payments.MethodDANA, ProviderName: "DANA", Active: true, SortOrder: 21},
		{MethodCode: payments.MethodOVO, ProviderName: "OVO", Active: true, SortOrder: 22},
		{MethodCode: payments.MethodShopeePay, ProviderName: "ShopeePay", Active: true, SortOrder: 23},
		{MethodCode: payments.MethodAlfamart, ProviderName: "Alfamart", Active: true, SortOrder: 30},
		{MethodCode: payments.MethodIndomaret, ProviderName: "Indomaret", Active: true, SortOrder: 31},
	}
	for i := range defaults {
		if err := db.WithContext(ctx).Create(&defaults[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedRoles(ctx context.Context, db *gorm.DB) error {
	roles := []role.Role{
		{Name: role.RoleAdmin, Description: "Full management access"},
		{Name: role.RoleManager, Description: "Monitoring, analytics, reports"},
		{Name: role.RolePetugas, Description: "Ticket verification at the gate"},
		{Name: role.RolePengunjung, Description: "Regular visitor account"},
		// ROLE-20: role super_admin wajib ada agar endpoint /super-admin/*
		// (manajemen penetapan role) punya pemilik yang sah.
		{Name: role.RoleSuperAdmin, Description: "Super administrator: manages role assignment"},
	}
	for _, r := range roles {
		var count int64
		if err := db.WithContext(ctx).Model(&role.Role{}).Where("name = ?", r.Name).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.WithContext(ctx).Create(&r).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type staffAccount struct {
	name        string
	email       string
	phone       string
	password    string
	passwordEnv string
	roleName    string
	gate        string
	shift       string
}

// userRow menyisipkan langsung ke users dan mengembalikan id yang dibuat.
type userRow struct {
	ID           string `gorm:"type:char(36);primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	IsActive     bool
	IsVerified   bool
	Gate         string
	Shift        string
}

func (userRow) TableName() string { return "users" }

const staffEmailDomain = "kawahputih.id"

func staffAccountsFromEnv() []staffAccount {
	configured := []staffAccount{
		{name: "Administrator Kawah Putih", email: "admin@" + staffEmailDomain, phone: "081200000002", passwordEnv: "SEED_ADMIN_PASSWORD", roleName: role.RoleAdmin},
		{name: "Manajer Kawah Putih", email: "manager@" + staffEmailDomain, phone: "081200000003", passwordEnv: "SEED_MANAGER_PASSWORD", roleName: role.RoleManager},
		{name: "Petugas Kawah Putih", email: "petugas@" + staffEmailDomain, phone: "081200000004", passwordEnv: "SEED_PETUGAS_PASSWORD", roleName: role.RolePetugas, gate: "Gerbang 1", shift: "Pagi (07:00 - 15:00)"},
		// ROLE-20: akun super_admin untuk mengelola penetapan role lewat /super-admin/*.
		{name: "Super Admin Kawah Putih", email: "superadmin@" + staffEmailDomain, phone: "081200000001", passwordEnv: "SEED_SUPERADMIN_PASSWORD", roleName: role.RoleSuperAdmin},
	}
	accounts := make([]staffAccount, 0, len(configured))
	for _, account := range configured {
		account.password = strings.TrimSpace(os.Getenv(account.passwordEnv))
		if account.password != "" {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

func seedUsers(ctx context.Context, db *gorm.DB) error {
	// Bersihkan akun admin legacy dari seed lama yang emailnya salah format
	// ("admin@kawahputih.idkawahputih.id") agar tidak mengganggu pembuatan akun yang benar.
	const brokenEmail = "admin@kawahputih.id" + staffEmailDomain
	if err := db.WithContext(ctx).Exec(
		`DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE email = ?)`, brokenEmail,
	).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).Exec(`DELETE FROM users WHERE email = ?`, brokenEmail).Error; err != nil {
		return err
	}

	for _, acc := range staffAccountsFromEnv() {
		var count int64
		if err := db.WithContext(ctx).Table("users").Where("email = ?", acc.email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(acc.password), 12)
		if err != nil {
			return err
		}
		row := userRow{
			ID: uuid.NewString(), Name: acc.name, Email: acc.email, Phone: acc.phone,
			PasswordHash: string(hash), IsActive: true, IsVerified: true,
			Gate: acc.gate, Shift: acc.shift,
		}
		if err := db.WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}
		// tautkan role
		var roleRow struct {
			ID string
		}
		if err := db.WithContext(ctx).Table("roles").Select("id").Where("name = ?", acc.roleName).Scan(&roleRow).Error; err != nil {
			return err
		}
		if roleRow.ID == "" {
			return fmt.Errorf("role %s missing", acc.roleName)
		}
		if err := db.WithContext(ctx).Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", row.ID, roleRow.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedTicketTypes(ctx context.Context, db *gorm.DB) error {
	type t struct {
		Code, Name, Description, Category string
		Price                             int64
		Capacity                          int
	}
	rows := []t{
		// HTM (Masuk)
		{"HTM-WNI-REG", "HTM Wisatawan Nusantara", "HTM wisatawan Nusantara (reguler).", "ENTRANCE", 31000, 1000},
		{"HTM-WNA-REG", "HTM Wisatawan Mancanegara", "HTM wisatawan mancanegara (reguler).", "ENTRANCE", 90000, 500},
		{"HTM-WNI-SUNRISE", "HTM Wisatawan Nusantara (Sunrise)", "HTM wisatawan Nusantara paket sunrise. Termasuk Sunan Ibu & Kawah Putih.", "SUNRISE", 46000, 500},
		{"HTM-WNA-SUNRISE", "HTM Wisatawan Mancanegara (Sunrise)", "HTM wisatawan mancanegara paket sunrise. Termasuk Sunan Ibu & Kawah Putih.", "SUNRISE", 105000, 300},

		// LHKP Charge
		{"LHKP-REG", "Charge LHKP (Reguler)", "Lintas Hutan Kawah Putih (LHKP).", "TRANSPORT", 179000, 200},
		{"LHKP-SUNRISE", "Charge LHKP (Sunrise)", "Lintas Hutan Kawah Putih (LHKP) untuk paket sunrise.", "TRANSPORT", 179000, 200},

		// Ontang Anting (Kuda)
		{"ONTANG-ANTING", "Ontang Anting", "Layanan ontang anting (kuda tunggang).", "ACTIVITY", 32000, 150},
		{"ONTANG-ANTING-SUNRISE", "Ontang Anting (Sunrise)", "Layanan ontang anting (kuda tunggang) untuk paket sunrise.", "ACTIVITY", 38000, 150},

		// Kendaraan
		{"VEHICLE-RODA2", "Kendaraan Roda 2", "Biaya masuk kendaraan roda 2.", "TRANSPORT", 10000, 1000},
		{"VEHICLE-RODA4", "Kendaraan Roda 4", "Biaya masuk kendaraan roda 4.", "TRANSPORT", 18000, 800},
		{"VEHICLE-RODA6", "Kendaraan Roda 6", "Biaya masuk kendaraan roda 6.", "TRANSPORT", 48000, 300},

		// Layanan
		{"HELM-STORAGE", "Jasa Penitipan Helm", "Jasa penitipan helm.", "SERVICE", 3000, 500},
		{"JASA-FOTOGRAFER", "Jasa Fotografer", "Layanan jasa fotografer.", "SERVICE", 11000, 50},

		// Aktivitas
		{"KUDA-TUNGGANG", "Kuda Tunggang", "Aktivitas naik kuda tunggangan.", "ACTIVITY", 55000, 80},
		{"PANAHAN", "Panahan", "Aktivitas panahan.", "ACTIVITY", 22000, 100},
		{"SUNAN-IBU", "Sunan Ibu", "Akses ke area Sunan Ibu.", "ACTIVITY", 15000, 300},
		{"JEMBATAN-APUNG", "Jembatan Apung", "Akses jembatan apung.", "ACTIVITY", 17000, 200},
		{"SKYWALK-CANOPY", "Skywalk Canopy", "Akses Skywalk Canopy.", "ACTIVITY", 17000, 200},

		// Paket/Layanan premium
		{"PREWEDDING", "Prewedding", "Layanan prewedding di Kawah Putih.", "SERVICE", 550000, 20},
		{"SHOOTING", "Shooting", "Layanan shooting/produser.", "SERVICE", 3300000, 10},
	}
	for _, r := range rows {
		// Upsert berdasarkan kode: sisipkan bila belum ada, selain itu segarkan nama/harga agar DB lama ikut tarif baru, lalu aktifkan lagi agar bisa dipesan.
		var existing tickettype.TicketType
		err := db.WithContext(ctx).Where("code = ?", r.Code).First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == gorm.ErrRecordNotFound {
			tt := tickettype.TicketType{
				Code: r.Code, Name: r.Name, Description: r.Description, Category: r.Category,
				Price: r.Price, Capacity: r.Capacity, IsActive: true,
			}
			if err := db.WithContext(ctx).Create(&tt).Error; err != nil {
				return err
			}
			continue
		}
		existing.Name = r.Name
		existing.Description = r.Description
		existing.Category = r.Category
		existing.Price = r.Price
		existing.Capacity = r.Capacity
		existing.IsActive = true
		if err := db.WithContext(ctx).Save(&existing).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedPackages(ctx context.Context, db *gorm.DB) error {
	type p struct {
		Name, Slug, Description string
		PriceCents              int64
		DurationHours           int
		MaxCapacity             int
	}
	rows := []p{
		{"Kawah Putih Standard", "kawah-putih-standard", "Tiket masuk + shuttle PP + welcome drink.", 45000, 3, 500},
		{"Sunrise Tour", "sunrise-tour", "Akses pagi (sebelum jam buka umum) + caldera tour + guide.", 95000, 5, 120},
		{"Family Weekend", "family-weekend", "Paket keluarga akhir pekan: masuk 2 dewasa + 3 anak + boat.", 120000, 6, 80},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("tourism_packages").Where("slug = ?", r.Slug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		pkg := tourismpackage.Package{
			Name: r.Name, Slug: r.Slug, Description: r.Description,
			PriceCents: r.PriceCents, Currency: "IDR", DurationHours: r.DurationHours,
			MaxCapacity: r.MaxCapacity, IsActive: true,
		}
		if err := db.WithContext(ctx).Create(&pkg).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedArticles(ctx context.Context, db *gorm.DB) error {
	type a struct {
		Title, Slug, Excerpt, Content string
	}
	rows := []a{
		{"Sejarah Kawah Putih", "sejarah-kawah-putih", "Dari kawah Gunung Patuha yang legendaris hingga destinasi wisata ikonik Jawa Barat.", "Kawah Putih ditemukan oleh seorang ahli kimia berkebangsaan Jerman, Dr. Franz Wilhelm Junghuhn, pada tahun 1837."},
		{"Tips Berkunjung ke Kawah Putih", "tips-berkunjung-kawah-putih", "Panduan singkat: jam terbaik, cuaca, dan perlengkapan yang perlu dibawa.", "Datanglah pada pagi hari sekitar pukul 07.00–10.00 agar kabut tipis belum turun dan pemandangan kawah paling jernih."},
		{"Fasilitas di Kawasan Wisata", "fasilitas-kawasan-wisata", "Area parkir, kamar mandi, mushola, dan penginapan di sekitar kawah.", "Kawasan wisata dilengkapi area parkir luas, mushola, kamar mandi, warung makan, dan spotting area foto."},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("articles").Where("slug = ?", r.Slug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		art := article.Article{
			Title: r.Title, Slug: r.Slug, Excerpt: r.Excerpt, Content: r.Content,
			Status: article.StatusPublished, AuthorID: "", PublishedAt: new(int64),
		}
		*art.PublishedAt = 0
		if err := db.WithContext(ctx).Create(&art).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedGallery(ctx context.Context, db *gorm.DB) error {
	type g struct {
		Title, ImageURL, Category string
		SortOrder                 int
	}
	rows := []g{
		{"Lautan Kawah Putih", "https://images.unsplash.com/photo-1571759005543-d05bfa88af9c", "kawah", 1},
		{"Kabut Pagi", "https://images.unsplash.com/photo-1544716278-ca5e3f4abd8c", "kawah", 2},
		{"Pulau Putih", "https://images.unsplash.com/photo-1505142468610-359e7d316be0", "pulau", 3},
		{"Hutan Rimba Patuha", "https://images.unsplash.com/photo-1441119805707-b3f5ef5bbc0a", "hutan", 4},
		{"Jalan Raya Ciwidey", "https://images.unsplash.com/photo-1469854523086-cc02fe5d8800", "transport", 5},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("gallery_items").Where("image_url = ?", r.ImageURL).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := gallery.Item{
			Title: r.Title, ImageURL: r.ImageURL, Category: r.Category, SortOrder: r.SortOrder, UploadedBy: "",
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedDestinationInfo(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Table("destination_information").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	item := destinationinfo.DestinationInfo{
		Title:           "Informasi Kawah Putih",
		Description:     "Kawah Putih adalah danau kawah vulkanik di kawasan Gunung Patuha, Ciwidey. Airnya berwarna putih kehijauan yang dapat berubah mengikuti kondisi alam, dikelilingi kabut tipis dan pepohonan.",
		History:         "Kawah Putih ditemukan oleh ahli kimia berkebangsaan Jerman, Dr. Franz Wilhelm Junghuhn, pada tahun 1837 saat meneliti Gunung Patuha.",
		DestinationDesc: "Destinasi alam unggulan Jawa Barat dengan panorama kaldera, jalur pejalan kaki, serta titik foto di bibir kawah.",
		Location:        "Ciwidey, Kabupaten Bandung, Jawa Barat",
		OperatingHours:  "07.00 - 17.00 WIB (setiap hari)",
		Contact:         "0800-1-123456",
		AdditionalInfo:  "Disarankan datang pada pagi hari untuk mendapatkan suasana kabut terbaik. Gunakan pakaian hangat karena suhu cukup dingin.",
		Status:          destinationinfo.StatusPublished,
	}
	return db.WithContext(ctx).Create(&item).Error
}

func seedFAQs(ctx context.Context, db *gorm.DB) error {
	type f struct {
		Question, Answer, Category string
		SortOrder                  int
	}
	rows := []f{
		{"Bagaimana cara membeli tiket?", "Pilih menu \"Pesan Tiket\", tentukan tanggal kunjungan dan jumlah tiket, lalu lanjutkan ke pembayaran. E-tiket dikirim setelah pembayaran dikonfirmasi.", "Tiket", 1},
		{"Jam operasional Kawah Putih?", "Kawasan dibuka setiap hari pukul 07.00 - 17.00 WIB. Disarankan datang pagi hari untuk suasana kabut terbaik.", "Jam Operasional", 2},
		{"Apa saja fasilitas yang tersedia?", "Tersedia area parkir, pusat informasi, musala, toilet, kafetaria, dan jalur pejalan kaki menuju kawah.", "Fasilitas", 3},
		{"Apakah tiket bisa dipesan online?", "Ya. Seluruh tiket dapat dipesan secara online melalui website ini dan dibayar melalui QRIS, transfer bank, atau kartu.", "Pembayaran", 4},
		{"Bagaimana cara melihat e-tiket?", "E-tiket tersedia di halaman E-Tiket setelah pembayaran. Anda dapat mengunduh, mencetak, atau membagikannya.", "Tiket", 5},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("faqs").Where("question = ?", r.Question).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := faq.FAQ{
			Question: r.Question, Answer: r.Answer, Category: r.Category,
			SortOrder: r.SortOrder, IsPublished: true,
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedFacilities(ctx context.Context, db *gorm.DB) error {
	type f struct {
		Name, Description, Icon, Location, OperatingHours string
		SortOrder                                         int
	}
	rows := []f{
		{"Area Parkir", "Area parkir luas untuk mobil pribadi dan bus wisata.", "Car", "Pintu Masuk Utama", "07.00 - 17.00", 1},
		{"Pusat Informasi", "Layanan informasi wisata dan pemandu kawasan.", "Info", "Area Penerimaan", "07.00 - 17.00", 2},
		{"Musala", "Tempat ibadah yang bersih untuk pengunjung.", "Landmark", "Area Fasilitas", "07.00 - 17.00", 3},
		{"Toilet & Kamar Bilas", "Toilet pria dan wanita beserta kamar bilas.", "DoorOpen", "Area Fasilitas", "07.00 - 17.00", 4},
		{"Kafetaria", "Makanan, minuman, dan camilan ringan.", "Coffee", "Area Wahana", "08.00 - 16.00", 5},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("facilities").Where("name = ?", r.Name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := facility.Facility{
			Name: r.Name, Description: r.Description, Icon: r.Icon, Location: r.Location,
			OperatingHours: r.OperatingHours, Status: facility.StatusActive,
			SortOrder: r.SortOrder, IsPublished: true,
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedVisitorRules(ctx context.Context, db *gorm.DB) error {
	type r struct {
		Title, Description, Category, Priority string
		SortOrder                              int
	}
	rows := []r{
		{"Tetap di Jalur yang Disediakan", "Demi keselamatan, pengunjung wajib berada di jalur pejalan kaki yang telah disediakan.", "Keselamatan", visitorrule.PriorityHigh, 1},
		{"Dilarang Berenang di Area Kawah", "Air kawah mengandung zat berbahaya dan area dilarang untuk aktivitas berenang.", "Larangan", visitorrule.PriorityCritical, 2},
		{"Jaga Kebersihan Kawasan", "Buanglah sampah pada tempatnya dan bawa turun sampah pribadi Anda.", "Kebersihan", visitorrule.PriorityNormal, 3},
		{"Patuhi Arahan Petugas", "Ikuti arahan petugas di lapangan untuk keselamatan dan kenyamanan bersama.", "Informasi Umum", visitorrule.PriorityNormal, 4},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("visitor_rules").Where("title = ?", r.Title).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := visitorrule.VisitorRule{
			Title: r.Title, Description: r.Description, Category: r.Category,
			Priority: r.Priority, SortOrder: r.SortOrder, IsPublished: true,
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}
