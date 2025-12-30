package yourpkg

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz" // MuPDF bindings for PDF->image
)

type InMemoryFile struct {
	Name string
	Data []byte
}

func (f *InMemoryFile) Filename() string { return f.Name }
func (f *InMemoryFile) Open() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.Data)), nil
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}
	t := mime.TypeByExtension(ext)
	return t != "" && strings.HasPrefix(t, "image/")
}

func isPDFFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}
	t := mime.TypeByExtension(ext)
	return t == "application/pdf"
}

// convertToImages converts an upload into one or more image pages.
func convertToImages(file *multipart.FileHeader) ([]*InMemoryFile, error) {
	if file == nil || file.Filename == "" {
		return nil, errors.New("filename not found")
	}

	// If already an image: just read and return it.
	if isImageFile(file.Filename) {
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		return []*InMemoryFile{{Name: file.Filename, Data: data}}, nil
	}

	// If PDF: render each page to JPEG.
	if isPDFFile(file.Filename) {
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		pdfBytes, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}

		doc, err := fitz.NewFromMemory(pdfBytes)
		if err != nil {
			return nil, fmt.Errorf("open pdf: %w", err)
		}
		defer doc.Close()

		n := doc.NumPage()
		out := make([]*InMemoryFile, 0, n)

		for i := 0; i < n; i++ {
			img, err := doc.Image(i)
			if err != nil {
				return nil, fmt.Errorf("render page %d: %w", i, err)
			}

			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
				return nil, fmt.Errorf("encode jpeg page %d: %w", i, err)
			}

			pageName := fmt.Sprintf("%s_%d.jpg", file.Filename, i)
			out = append(out, &InMemoryFile{Name: pageName, Data: buf.Bytes()})
		}

		return out, nil
	}

	return nil, errors.New("file type is not supported")
}
