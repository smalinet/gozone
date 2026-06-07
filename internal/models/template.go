package models

// ZoneTemplate represents a reusable DNS record template.
type ZoneTemplate struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsBuiltin   bool   `json:"is_builtin"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ZoneTemplateRecord is a single record within a zone template.
type ZoneTemplateRecord struct {
	ID         int64  `json:"id"`
	TemplateID int64  `json:"template_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
	Priority   int    `json:"priority"`
	Disabled   bool   `json:"disabled"`
}
