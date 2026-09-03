package websetting

type UpsertItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
	Type  string `json:"type" binding:"omitempty,oneof=string text json image"`
}

type UpsertRequest struct {
	Settings []UpsertItem `json:"settings" binding:"required,dive"`
}

type Response struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
