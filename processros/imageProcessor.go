package processros

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Xenos112/go-cdn/constants"
	"github.com/chai2010/webp"
	"github.com/google/uuid"
)

func ProcessFileUpload(header *multipart.FileHeader, host string) constants.UploadResult {
	result := constants.UploadResult{FileName: header.Filename}

	file, err := header.Open()
	if err != nil {
		result.Error = "Failed to open file"
		return result
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		result.Error = "Failed to read file"
		return result
	}
	contentType := http.DetectContentType(buf[:n])

	if seeker, ok := file.(io.Seeker); ok {
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			result.Error = "File processing error"
			return result
		}
	} else {
		result.Error = "File type not supported"
		return result
	}

	if strings.HasPrefix(contentType, "image/") && contentType != "image/webp" {
		var img image.Image
		var decodeErr error

		switch contentType {
		case "image/jpeg":
			img, decodeErr = jpeg.Decode(file)
		case "image/png":
			img, decodeErr = png.Decode(file)
		case "image/gif":
			img, decodeErr = gif.Decode(file)
		default:
		}

		if decodeErr == nil && img != nil {
			newFileName := uuid.New().String() + ".webp"
			filePath := filepath.Join(constants.UploadDir, newFileName)

			dst, err := os.Create(filePath)
			if err != nil {
				result.Error = "Failed to create file"
				return result
			}
			defer dst.Close()

			if err := webp.Encode(dst, img, &webp.Options{Quality: 80}); err != nil {
				result.Error = "Failed to convert image"
				return result
			}

			result.URL = fmt.Sprintf("http://%s/files/%s", host, newFileName)
			return result
		}
	}

	if seeker, ok := file.(io.Seeker); ok {
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			result.Error = "File processing error"
			return result
		}
	} else {
		result.Error = "File type not supported"
		return result
	}

	fileExt := filepath.Ext(header.Filename)
	newFileName := uuid.New().String() + fileExt
	filePath := filepath.Join(constants.UploadDir, newFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		result.Error = "Failed to create file"
		return result
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		result.Error = "Failed to save file"
		return result
	}

	result.URL = fmt.Sprintf("http://%s/files/%s", host, newFileName)
	return result
}
