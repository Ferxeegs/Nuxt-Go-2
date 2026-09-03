package media

type Response struct {
	ID            uint   `json:"id"`
	FileName      string `json:"file_name"`
	Path          string `json:"path"`
	URL           string `json:"url"`
	MimeType      string `json:"mime_type"`
	Size          int64  `json:"size"`
	RelatedModule string `json:"related_module,omitempty"`
	CreatedAt     string `json:"created_at"`
}
