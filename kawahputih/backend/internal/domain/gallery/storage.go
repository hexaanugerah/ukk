package gallery

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// newImageFileName menghasilkan nama berkas acak yang aman.
func newImageFileName(ext string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("gallery: could not generate file name: %v", err))
	}
	return hex.EncodeToString(buf) + ext
}

func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("gallery: create upload dir: %w", err)
	}
	return nil
}

func writeFile(target string, src io.Reader) error {
	f, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("gallery: create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		return fmt.Errorf("gallery: write file: %w", err)
	}
	return nil
}
