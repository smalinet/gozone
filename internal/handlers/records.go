package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/babykart/gozone/internal/logger"
	"github.com/babykart/gozone/internal/middleware"
	"github.com/babykart/gozone/internal/models"
	"github.com/babykart/gozone/internal/validators"
)

// CreateRecordPage renders the record creation form for a zone (GET /zones/{zone_id}/records/new).
func (h *Handler) CreateRecordPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	zoneID := r.PathValue("zone_id")

	zone, err := h.PDNS.GetZone(r.Context(), zoneID)
	if err != nil {
		h.renderError(w, r, "Zone not found")
		return
	}

	data := map[string]interface{}{
		"Title":       "Add Record - " + zone.Name + " - GoZone",
		"User":        user,
		"Zone":        zone,
		"RecordTypes": GetRecordTypes(),
	}
	h.render(w, r, "record_create.html", data)
}

// CreateRecord creates a DNS record in a zone from form data (POST /zones/{zone_id}/records/create).
//
// Accepts name, type, content, ttl, and priority form values. Defaults TTL to 3600.
func (h *Handler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	zoneID := r.PathValue("zone_id")

	name := strings.TrimSpace(r.FormValue("name"))
	recordType := strings.TrimSpace(r.FormValue("type"))
	content := strings.TrimSpace(r.FormValue("content"))
	ttlStr := strings.TrimSpace(r.FormValue("ttl"))
	priorityStr := strings.TrimSpace(r.FormValue("priority"))

	ttl, err := strconv.Atoi(ttlStr)
	if err != nil || ttl <= 0 {
		ttl = 3600
	}

	priority := 0
	if priorityStr != "" {
		priority, _ = strconv.Atoi(priorityStr)
	}

	if name == "" || recordType == "" || content == "" {
		// #nosec G710 -- zoneID from chi r.PathValue, controlled by route pattern
		http.Redirect(w, r, "/zones/"+zoneID+"/records/new", http.StatusSeeOther)
		return
	}

	name = normalizeRecordName(name, zoneID)

	if err := validators.ValidateRecordType(recordType); err != nil {
		h.renderError(w, r, "Invalid record type: "+err.Error())
		return
	}

	if err := validators.ValidateRecordContent(recordType, content); err != nil {
		h.renderError(w, r, "Invalid record content: "+err.Error())
		return
	}

	recordContent, recordPriority := prepareMXSRVContent(recordType, content, priority)

	rrset := models.RRSet{
		Name: name,
		Type: recordType,
		TTL:  ttl,
		Records: []models.RecordInfo{
			{
				Content:  recordContent,
				Priority: recordPriority,
				Disabled: false,
			},
		},
	}

	if err := h.PDNS.CreateRecord(r.Context(), zoneID, rrset); err != nil {
		h.renderInternalError(w, r, "Failed to create record", err)
		return
	}

	if _, err := h.DB.Exec(
		"INSERT INTO activity_logs (user_id, zone_id, action, details) VALUES (?, ?, 'create_record', ?)",
		user.ID, zoneID, fmt.Sprintf("Created %s record %s -> %s", recordType, name, content),
	); err != nil {
		logger.Error("failed to log create_record activity", "zone_id", zoneID, "error", err)
	}

	// #nosec G710 -- zoneID from chi r.PathValue, controlled by route pattern
	http.Redirect(w, r, "/zones/"+zoneID, http.StatusSeeOther)
}

// EditRecordPage renders the record edit form (GET /zones/{zone_id}/records/edit?name=...&type=...).
//
// The record to edit is identified by the "name" and "type" query parameters.
func (h *Handler) EditRecordPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	zoneID := r.PathValue("zone_id")
	recordQuery := r.URL.Query()

	recordName := recordQuery.Get("name")
	recordType := recordQuery.Get("type")

	zone, err := h.PDNS.GetZone(r.Context(), zoneID)
	if err != nil {
		h.renderError(w, r, "Zone not found")
		return
	}

	records, err := h.PDNS.ListRecords(r.Context(), zoneID)
	if err != nil {
		h.renderError(w, r, "Failed to fetch records")
		return
	}

	var targetRRSet *models.RRSet
	for _, rr := range records {
		if rr.Name == recordName && rr.Type == recordType {
			targetRRSet = &rr
			break
		}
	}

	if targetRRSet == nil {
		h.renderError(w, r, "Record not found")
		return
	}

	data := map[string]interface{}{
		"Title":       "Edit Record - " + zone.Name + " - GoZone",
		"User":        user,
		"Zone":        zone,
		"Record":      targetRRSet,
		"RecordTypes": GetRecordTypes(),
	}
	h.render(w, r, "record_edit.html", data)
}

// UpdateRecord replaces a DNS record in a zone from form data (POST /zones/{zone_id}/records/update).
//
// Uses the REPLACE changetype to ensure idempotent updates. Accepts name, type,
// content, ttl, priority, and disabled form values.
func (h *Handler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	zoneID := r.PathValue("zone_id")

	name, recordType, content, ttl, priority, disabled, err := parseRecordForm(r)
	if err != nil {
		h.renderInternalError(w, r, "Invalid record form", err)
		return
	}
	if name == "" || recordType == "" || content == "" {
		h.renderError(w, r, "Name, type, and content are required")
		return
	}

	name = normalizeRecordName(name, zoneID)

	if err := validators.ValidateRecordType(recordType); err != nil {
		h.renderError(w, r, "Invalid record type: "+err.Error())
		return
	}

	if err := validators.ValidateRecordContent(recordType, content); err != nil {
		h.renderError(w, r, "Invalid record content: "+err.Error())
		return
	}

	recordContent, recordPriority := prepareMXSRVContent(recordType, content, priority)

	rrset := models.RRSet{
		Name: name,
		Type: recordType,
		TTL:  ttl,
		Records: []models.RecordInfo{
			{
				Content:  recordContent,
				Priority: recordPriority,
				Disabled: disabled,
			},
		},
	}

	if err := h.PDNS.UpdateRecord(r.Context(), zoneID, rrset); err != nil {
		h.renderInternalError(w, r, "Failed to update record", err)
		return
	}

	user := middleware.GetUser(r)
	if _, err := h.DB.Exec(
		"INSERT INTO activity_logs (user_id, zone_id, action, details) VALUES (?, ?, 'update_record', ?)",
		user.ID, zoneID, fmt.Sprintf("Updated %s record %s", recordType, name),
	); err != nil {
		logger.Error("failed to log update_record activity", "zone_id", zoneID, "error", err)
	}

	// #nosec G710 -- zoneID from chi r.PathValue, controlled by route pattern
	http.Redirect(w, r, "/zones/"+zoneID, http.StatusSeeOther)
}

// InlineUpdateRecord updates a record via AJAX and returns JSON (POST /zones/{zone_id}/records/inline-update).
func (h *Handler) InlineUpdateRecord(w http.ResponseWriter, r *http.Request) {
	zoneID := r.PathValue("zone_id")

	name, recordType, content, ttl, priority, disabled, err := parseRecordForm(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if name == "" || recordType == "" || content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name, type, and content are required"})
		return
	}

	name = normalizeRecordName(name, zoneID)

	if err := validators.ValidateRecordType(recordType); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid record type: " + err.Error()})
		return
	}

	if err := validators.ValidateRecordContent(recordType, content); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid record content: " + err.Error()})
		return
	}

	recordContent, recordPriority := prepareMXSRVContent(recordType, content, priority)

	rrset := models.RRSet{
		Name: name,
		Type: recordType,
		TTL:  ttl,
		Records: []models.RecordInfo{
			{
				Content:  recordContent,
				Priority: recordPriority,
				Disabled: disabled,
			},
		},
	}

	if err := h.PDNS.UpdateRecord(r.Context(), zoneID, rrset); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update record"})
		return
	}

	user := middleware.GetUser(r)
	if _, err := h.DB.Exec(
		"INSERT INTO activity_logs (user_id, zone_id, action, details) VALUES (?, ?, 'update_record', ?)",
		user.ID, zoneID, fmt.Sprintf("Updated %s record %s", recordType, name),
	); err != nil {
		logger.Error("failed to log update_record activity", "zone_id", zoneID, "error", err)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"record":  rrset,
	})
}

// BatchCreateRecords creates multiple DNS records in a zone (POST /zones/{zone_id}/records/batch-create).
func (h *Handler) BatchCreateRecords(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	zoneID := r.PathValue("zone_id")

	if err := r.ParseForm(); err != nil {
		// #nosec G710 -- zoneID from chi r.PathValue, controlled by route pattern
		http.Redirect(w, r, "/zones/"+zoneID+"/records/new", http.StatusSeeOther)
		return
	}

	names := r.PostForm["name"]
	types := r.PostForm["type"]
	contents := r.PostForm["content"]
	ttls := r.PostForm["ttl"]
	priorities := r.PostForm["priority"]

	if len(names) == 0 || len(types) == 0 || len(contents) == 0 {
		h.renderError(w, r, "At least one record is required")
		return
	}

	type logEntry struct {
		recordType string
		name       string
		content    string
	}

	var rrsets []models.RRSet
	var logEntries []logEntry
	for i := 0; i < len(names); i++ {
		name := strings.TrimSpace(names[i])
		recordType := strings.TrimSpace(types[i])
		content := strings.TrimSpace(contents[i])

		if name == "" || recordType == "" || content == "" {
			continue
		}

		name = normalizeRecordName(name, zoneID)

		ttl := 3600
		if i < len(ttls) {
			if t, err := strconv.Atoi(strings.TrimSpace(ttls[i])); err == nil && t > 0 {
				ttl = t
			}
		}
		priority := 0
		if i < len(priorities) {
			if p, err := strconv.Atoi(strings.TrimSpace(priorities[i])); err == nil && p > 0 {
				priority = p
			}
		}

		if err := validators.ValidateRecordType(recordType); err != nil {
			h.renderError(w, r, "Invalid record type '"+recordType+"': "+err.Error())
			return
		}
		if err := validators.ValidateRecordContent(recordType, content); err != nil {
			h.renderError(w, r, "Invalid record content: "+err.Error())
			return
		}

		recordContent, recordPriority := prepareMXSRVContent(recordType, content, priority)

		rrsets = append(rrsets, models.RRSet{
			Name: name,
			Type: recordType,
			TTL:  ttl,
			Records: []models.RecordInfo{
				{Content: recordContent, Priority: recordPriority, Disabled: false},
			},
		})
		logEntries = append(logEntries, logEntry{recordType, name, content})
	}

	if len(rrsets) == 0 {
		h.renderError(w, r, "No valid records to create")
		return
	}

	if err := h.PDNS.CreateRecords(r.Context(), zoneID, rrsets); err != nil {
		h.renderInternalError(w, r, "Failed to create records", err)
		return
	}

	for _, e := range logEntries {
		if _, err := h.DB.Exec(
			"INSERT INTO activity_logs (user_id, zone_id, action, details) VALUES (?, ?, 'create_record', ?)",
			user.ID, zoneID, fmt.Sprintf("Created %s record %s -> %s", e.recordType, e.name, e.content),
		); err != nil {
			logger.Error("failed to log create_record activity", "zone_id", zoneID, "error", err)
		}
	}

	// #nosec G710 -- zoneID from chi r.PathValue, controlled by route pattern
	http.Redirect(w, r, "/zones/"+zoneID, http.StatusSeeOther)
}

// prepareMXSRVContent embeds priority in the content field for MX and SRV records.
// It strips any existing priority prefix from content first when the token count
// indicates the content was read from PowerDNS (which stores priority directly in
// the content field). For MX: 2+ tokens with leading number = PDNS read path.
// For SRV: 4+ tokens with leading number = PDNS read path (priority weight port target).
// Always returns 0 for the RecordInfo priority because PowerDNS rejects the
// "priority" element in record PATCH body.
func prepareMXSRVContent(recordType, content string, priority int) (string, int) {
	if recordType != "MX" && recordType != "SRV" {
		return content, priority
	}
	tokens := strings.Fields(content)
	if len(tokens) > 0 {
		if _, err := strconv.Atoi(tokens[0]); err == nil {
			// Only strip when token count indicates PDNS read-path content:
			// MX from PDNS: "10 mail.example.com." (2+ tokens)
			// SRV from PDNS: "10 5 5060 target." (4 tokens: priority weight port target)
			// SRV from form:  "5 5060 target."    (3 tokens: weight port target)
			isPDNSFormat := (recordType == "MX" && len(tokens) >= 2) ||
				(recordType == "SRV" && len(tokens) >= 4)
			if isPDNSFormat {
				content = strings.Join(tokens[1:], " ")
			}
		}
	}
	return fmt.Sprintf("%d %s", priority, content), 0
}

// normalizeRecordName cleans a user-supplied record name for the PDNS PATCH API.
// PDNS expects names relative to the zone (e.g., "www") or the zone name itself
// for apex records. Names ending with the zone suffix are stripped back to their
// relative form. "@" is mapped to the zone name.
func normalizeRecordName(name, zoneName string) string {
	name = strings.TrimSpace(name)
	if name == "@" {
		return zoneName
	}
	zone := strings.TrimSuffix(zoneName, ".")
	if strings.HasSuffix(name, "."+zone) || strings.HasSuffix(name, "."+zone+".") {
		name = strings.TrimSuffix(name, "."+zone+".")
		name = strings.TrimSuffix(name, "."+zone)
	} else if strings.EqualFold(name, zone) || strings.EqualFold(name, zoneName) {
		name = zoneName
	}
	if name == "" {
		return zoneName
	}
	return name
}

func parseRecordForm(r *http.Request) (name, recordType, content string, ttl, priority int, disabled bool, err error) {
	name = strings.TrimSpace(r.FormValue("name"))
	recordType = strings.TrimSpace(r.FormValue("type"))
	content = strings.TrimSpace(r.FormValue("content"))
	ttlStr := strings.TrimSpace(r.FormValue("ttl"))
	priorityStr := strings.TrimSpace(r.FormValue("priority"))
	disabled = r.FormValue("disabled") == "on" || r.FormValue("disabled") == "true"

	ttl, err = strconv.Atoi(ttlStr)
	if err != nil || ttl <= 0 {
		ttl = 3600
		err = nil
	}

	priority = 0
	if priorityStr != "" {
		priority, _ = strconv.Atoi(priorityStr)
	}
	return
}

// DeleteRecord deletes a DNS record from a zone (POST /zones/{zone_id}/records/delete).
//
// Identifies the record by "name" and "type" form values.
func (h *Handler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	zoneID := r.PathValue("zone_id")

	recordName := r.FormValue("name")
	recordType := r.FormValue("type")

	if err := h.PDNS.DeleteRecord(r.Context(), zoneID, recordName, recordType); err != nil {
		h.renderInternalError(w, r, "Failed to delete record", err)
		return
	}

	if _, err := h.DB.Exec(
		"INSERT INTO activity_logs (user_id, zone_id, action, details) VALUES (?, ?, 'delete_record', ?)",
		user.ID, zoneID, fmt.Sprintf("Deleted %s record %s", recordType, recordName),
	); err != nil {
		logger.Error("failed to log delete_record activity", "zone_id", zoneID, "error", err)
	}

	// #nosec G710 -- zoneID from chi r.PathValue, controlled by route pattern
	http.Redirect(w, r, "/zones/"+zoneID, http.StatusSeeOther)
}
