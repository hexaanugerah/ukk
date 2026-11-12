-- =============================================================================
-- Reference/demo seed data for the Kawah Putih booking schema.
-- Idempotent: re-running never duplicates rows.
-- =============================================================================

-- Application roles (must match the backend's role constants and the
-- frontend's AppRole union).
INSERT INTO roles (id, name) VALUES
    (1, 'ADMIN'),
    (2, 'MANAGER'),
    (3, 'PETUGAS'),
    (4, 'PENGUNJUNG')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Bookable ticket catalog — the official "Price List Wana Wisata Kawah
-- Putih". Keep in sync with backend/internal/seed/seed.go, which is the
-- authoritative seeder for the Go runtime.
-- (quota mirrors the reference capacity; nothing is sold yet, so
-- available_quota starts equal to quota).
INSERT INTO tickets (id, name, description, price, quota, available_quota, status) VALUES
    (1,  'HTM Wisatawan Nusantara',                 'HTM wisatawan Nusantara (reguler).',                                31000.00, 1000, 1000, 'AKTIF'),
    (2,  'HTM Wisatawan Mancanegara',               'HTM wisatawan mancanegara (reguler).',                              90000.00,  500,  500, 'AKTIF'),
    (3,  'HTM Wisatawan Nusantara (Sunrise)',       'HTM wisatawan Nusantara paket sunrise. Termasuk Sunan Ibu & Kawah Putih.', 46000.00, 500, 500, 'AKTIF'),
    (4,  'HTM Wisatawan Mancanegara (Sunrise)',     'HTM wisatawan mancanegara paket sunrise. Termasuk Sunan Ibu & Kawah Putih.', 105000.00, 300, 300, 'AKTIF'),
    (5,  'Charge LHKP (Reguler)',                   'Lintas Hutan Kawah Putih (LHKP).',                                 179000.00,  200,  200, 'AKTIF'),
    (6,  'Charge LHKP (Sunrise)',                   'Lintas Hutan Kawah Putih (LHKP) untuk paket sunrise.',             179000.00,  200,  200, 'AKTIF'),
    (7,  'Ontang Anting',                           'Layanan ontang anting (kuda tunggang).',                            32000.00,  150,  150, 'AKTIF'),
    (8,  'Ontang Anting (Sunrise)',                 'Layanan ontang anting (kuda tunggang) untuk paket sunrise.',        38000.00,  150,  150, 'AKTIF'),
    (9,  'Kendaraan Roda 2',                        'Biaya masuk kendaraan roda 2.',                                     10000.00, 1000, 1000, 'AKTIF'),
    (10, 'Kendaraan Roda 4',                        'Biaya masuk kendaraan roda 4.',                                     18000.00,  800,  800, 'AKTIF'),
    (11, 'Kendaraan Roda 6',                        'Biaya masuk kendaraan roda 6.',                                     48000.00,  300,  300, 'AKTIF'),
    (12, 'Jasa Penitipan Helm',                     'Jasa penitipan helm.',                                               3000.00,  500,  500, 'AKTIF'),
    (13, 'Jasa Fotografer',                         'Layanan jasa fotografer.',                                          11000.00,   50,   50, 'AKTIF'),
    (14, 'Kuda Tunggang',                           'Aktivitas naik kuda tunggangan.',                                   55000.00,   80,   80, 'AKTIF'),
    (15, 'Panahan',                                 'Aktivitas panahan.',                                                22000.00,  100,  100, 'AKTIF'),
    (16, 'Sunan Ibu',                               'Akses ke area Sunan Ibu.',                                          15000.00,  300,  300, 'AKTIF'),
    (17, 'Jembatan Apung',                          'Akses jembatan apung.',                                             17000.00,  200,  200, 'AKTIF'),
    (18, 'Skywalk Canopy',                          'Akses Skywalk Canopy.',                                             17000.00,  200,  200, 'AKTIF'),
    (19, 'Prewedding',                              'Layanan prewedding di Kawah Putih.',                               550000.00,   20,   20, 'AKTIF'),
    (20, 'Shooting',                                'Layanan shooting/produser.',                                      3300000.00,   10,   10, 'AKTIF')
ON DUPLICATE KEY UPDATE name = VALUES(name);
