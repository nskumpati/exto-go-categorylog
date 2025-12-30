package yourpkg

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
)

// DocumentToBase64 converts a document (pdf, docx, etc.) into base64-encoded PNGs,
// returning a "data:{mime};base64,{...}" string per processed page.
//
// - maxPagesToProcess: nil means "process all pages"
// - includeFilename: if true, each item is a map {"base64": "...", "fileName": originalName}
//
// Return type is []any because Go doesn't have Union[str, Dict] types.
func DocumentToBase64(
	filePath *multipart.FileHeader,
	maxPagesToProcess *int,
	includeFilename bool,
) ([]any, error) {

	var base64Images []any

	files, err := shared.convertToImages(filePath)
	if err != nil {
		return nil, err
	}

	pageCount := len(files)

	pagesToProcess := pageCount
	if maxPagesToProcess != nil && *maxPagesToProcess > 0 {
		if *maxPagesToProcess < pageCount {
			pagesToProcess = *maxPagesToProcess
		}
	}

	for currPage := 0; currPage < pageCount && currPage < pagesToProcess; currPage++ {
		pageFile := files[currPage]

		// Default mime type
		mimeType := "image/png"

		// Guess mime based on extension if filename is present
		if pageFile != nil && pageFile.Filename != "" {
			ext := filepath.Ext(pageFile.Filename)
			if guessed := mime.TypeByExtension(ext); guessed != "" {
				mimeType = guessed
			}
		}

		// Open and read page bytes
		f, err := pageFile.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		base64String := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

		if includeFilename {
			base64Images = append(base64Images, map[string]string{
				"base64":   base64String,
				"fileName": filePath.Filename,
			})
		} else {
			base64Images = append(base64Images, base64String)
		}
	}

	return base64Images, nil
}

// BytesToHumanReadable converts a byte count to a human-readable string.
// - nil or 0 => "0 B"
// - negative => "Invalid"
// - units: B, KB, MB, GB, TB, PB (base-1024)
// - formatting: show 1 decimal if <10 and not integer, else no decimals.
func BytesToHumanReadable(numBytes *int64) string {
	if numBytes == nil || *numBytes == 0 {
		return "0 B"
	}
	if *numBytes < 0 {
		return "Invalid"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}

	// Find exponent similar to bit_length()//10
	// We advance while dividing by 1024 fits and units remain.
	exponent := 0
	value := float64(*numBytes)
	for exponent < len(units)-1 && value >= 1024.0 {
		value /= 1024.0
		exponent++
	}

	// Formatting rules
	var valueStr string
	if value >= 10.0 || value == float64(int64(value)) {
		valueStr = fmt.Sprintf("%.0f", value)
	} else {
		valueStr = fmt.Sprintf("%.1f", value)
	}

	return fmt.Sprintf("%s %s", valueStr, units[exponent])
}
