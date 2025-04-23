package constants

const (
	UploadDir     = "./uploads"
	MaxUploadSize = 100 * 1024 * 1024 // 100MB
)

type UploadResult struct {
	FileName string `json:"fileName"`
	URL      string `json:"url"`
	Error    string `json:"error,omitempty"`
}
