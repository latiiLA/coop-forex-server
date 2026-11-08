package response

type Status struct {
	IsSuccessful bool        `json:"is_successful"`
	Message      string      `json:"message,omitempty"`
	Error        string      `json:"error,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Meta         interface{} `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Scope string `json:"scope,omitempty"` // bank|org — set by queue endpoint
}

