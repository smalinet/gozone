package models

import (
	"encoding/json"
	"testing"
)

func TestZoneTemplate_JSON(t *testing.T) {
	tmpl := ZoneTemplate{
		ID:          1,
		Name:        "web",
		Description: "SOA + NS + A/AAAA + CNAME",
		IsBuiltin:   true,
		CreatedAt:   "2024-01-01 00:00:00",
		UpdatedAt:   "2024-01-01 00:00:00",
	}

	data, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatalf("marshal ZoneTemplate: %v", err)
	}

	var out ZoneTemplate
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal ZoneTemplate: %v", err)
	}

	if out.ID != 1 || out.Name != "web" || out.Description != "SOA + NS + A/AAAA + CNAME" || !out.IsBuiltin {
		t.Errorf("ZoneTemplate round-trip mismatch: %+v", out)
	}
}

func TestZoneTemplateRecord_JSON(t *testing.T) {
	rec := ZoneTemplateRecord{
		ID:         1,
		TemplateID: 2,
		Name:       "@",
		Type:       "A",
		Content:    "{{IP}}",
		TTL:        3600,
		Priority:   0,
		Disabled:   false,
	}

	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal ZoneTemplateRecord: %v", err)
	}

	var out ZoneTemplateRecord
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal ZoneTemplateRecord: %v", err)
	}

	if out.Name != "@" || out.Type != "A" || out.Content != "{{IP}}" || out.TTL != 3600 {
		t.Errorf("ZoneTemplateRecord round-trip mismatch: %+v", out)
	}
}
