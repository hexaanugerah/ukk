// Command worker masih placeholder. Belum ada job async yang butuh proses worker terpisah karena efek samping booking (email saat pembayaran berhasil) masih sinkron. Biner ini baru berguna setelah ada queue seperti Redis Streams atau RabbitMQ.
package main

import "fmt"

func main() {
	fmt.Println("KawahPutih worker: no queued jobs registered yet — nothing to do")
}
