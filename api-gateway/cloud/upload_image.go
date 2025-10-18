// Package cloud provides a simple fake cloud service for storing images.
// It can be used instead of a CDN or Cloudflare.
package cloud

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/binary_file_uploads/
func HandleBinaryFileUpload(w http.ResponseWriter, r *http.Request, params map[string]string) {

	if err := r.ParseMultipartForm(1 << 20); err != nil { // 1MB limit
		http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// helper to save each file
	saveFile := func(fieldName string) error {
		file, header, err := r.FormFile(fieldName)
		if err != nil {
			// if field is missing, skip silently
			if err == http.ErrMissingFile {
				return nil
			}
			return fmt.Errorf("failed to get file '%s': %w", fieldName, err)
		}
		defer file.Close()

		dstPath := filepath.Join("/images", header.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("failed to create file '%s': %w", dstPath, err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return fmt.Errorf("failed to save file '%s': %w", header.Filename, err)
		}

		return nil
	}

	if err := saveFile("merchant"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := saveFile("item"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("files uploaded successfully"))
}
