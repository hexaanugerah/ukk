// Command migration hanya pelengkap struktur; proyek memakai file SQL bernomor di database/migrations/*.up.sql (bukan runner migrasi Go). Biner ini sengaja no-op.
package main

import "fmt"

func main() {
	fmt.Println("Migrations are plain SQL in database/migrations/ — run ./scripts/migrate.sh instead of this binary.")
}
