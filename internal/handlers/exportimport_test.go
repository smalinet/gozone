package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/babykart/gozone/internal/models"
	"github.com/babykart/gozone/internal/testutil"
)

func testExportPDNS() testutil.PDNSHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		if strings.Contains(r.URL.RawQuery, "rrsets") {
			w.Write([]byte(`[{"name":"example.com.","id":"example.com.","kind":"Native","serial":2024010100}]`))
			return
		}

		if strings.Contains(path, "/zones/") && !strings.Contains(path, "/export") && !strings.Contains(path, "/import") && !strings.Contains(path, "/records") && !strings.Contains(path, "/cryptokeys") && !strings.Contains(path, "/metadata") {
			if r.Method == http.MethodGet {
				w.Write([]byte(`{"id":"example.com.","name":"example.com.","kind":"Native","serial":2024010100,"rrsets":[{"name":"example.com.","type":"SOA","ttl":3600,"records":[{"content":"ns1.example.com. hostmaster.example.com. 2024010100 3600 900 1209600 3600","disabled":false}]},{"name":"example.com.","type":"NS","ttl":3600,"records":[{"content":"ns1.example.com.","disabled":false}]},{"name":"www.example.com.","type":"A","ttl":3600,"records":[{"content":"192.0.2.1","disabled":false}]},{"name":"example.com.","type":"MX","ttl":3600,"records":[{"content":"mail.example.com.","disabled":false,"priority":10}]}]}`))
				return
			}
		}

		w.Write([]byte(`[]`))
	}
}

func TestExportZone_BIND(t *testing.T) {
	h, srv := newTestHandlerWithPDNS(t, testExportPDNS())
	defer srv.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/zones/example.com./export?format=bind", nil)
	r.SetPathValue("zone_id", "example.com.")
	r = withUserContext(r, &models.User{ID: 1, Username: "test", Role: "admin"})
	h.ExportZone(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.HasPrefix(body, "$ORIGIN example.com.") {
		t.Errorf("expected $ORIGIN, got: %s", body[:60])
	}
	if !strings.Contains(body, "$TTL") {
		t.Errorf("expected $TTL, got: %s", body[:100])
	}
	if !strings.Contains(body, "IN SOA") {
		t.Errorf("expected SOA record, got: %s", body)
	}
	if !strings.Contains(body, "IN NS") {
		t.Errorf("expected NS record, got: %s", body)
	}
	if !strings.Contains(body, "IN A") {
		t.Errorf("expected A record, got: %s", body)
	}
	if !strings.Contains(body, "IN MX") {
		t.Errorf("expected MX record, got: %s", body)
	}
	if !strings.Contains(body, "10 mail.example.com.") {
		t.Errorf("expected MX priority+content, got: %s", body)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain Content-Type, got: %s", ct)
	}
	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") {
		t.Errorf("expected Content-Disposition attachment, got: %s", cd)
	}
}

func TestExportZone_CSV(t *testing.T) {
	h, srv := newTestHandlerWithPDNS(t, testExportPDNS())
	defer srv.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/zones/example.com./export?format=csv", nil)
	r.SetPathValue("zone_id", "example.com.")
	r = withUserContext(r, &models.User{ID: 1, Username: "test", Role: "admin"})
	h.ExportZone(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "name,type,content,ttl,priority,disabled") {
		t.Errorf("expected CSV header, got: %s", body)
	}
	if !strings.Contains(body, "SOA") {
		t.Errorf("expected SOA in CSV, got: %s", body)
	}
	if !strings.Contains(body, "192.0.2.1") {
		t.Errorf("expected A record content in CSV, got: %s", body)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Errorf("expected text/csv Content-Type, got: %s", ct)
	}
}

func TestExportZone_InvalidFormat(t *testing.T) {
	h, srv := newTestHandlerWithPDNS(t, testExportPDNS())
	defer srv.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/zones/example.com./export?format=json", nil)
	r.SetPathValue("zone_id", "example.com.")
	r = withUserContext(r, &models.User{ID: 1, Username: "test", Role: "admin"})
	h.ExportZone(w, r)

	if !strings.Contains(w.Body.String(), "Invalid format") {
		t.Errorf("expected error message, got: %s", w.Body.String())
	}
}

func TestRelativeBindName(t *testing.T) {
	tests := []struct {
		name, origin, expected string
	}{
		{"example.com.", "example.com.", "@"},
		{"www.example.com.", "example.com.", "www"},
		{"test.www.example.com.", "example.com.", "test.www"},
		{"example.com.", "other.com.", "example.com."},
		{"example.com", "example.com.", "@"},
		{"www.example.com", "example.com.", "www"},
		{"other.com.", "example.com.", "other.com."},
	}

	for _, tc := range tests {
		t.Run(tc.name+"-"+tc.origin, func(t *testing.T) {
			result := relativeBindName(tc.name, tc.origin)
			if result != tc.expected {
				t.Errorf("relativeBindName(%q, %q) = %q, want %q", tc.name, tc.origin, result, tc.expected)
			}
		})
	}
}

func TestSortRRSets(t *testing.T) {
	records := []models.RRSet{
		{Name: "example.com.", Type: "A"},
		{Name: "example.com.", Type: "NS"},
		{Name: "example.com.", Type: "SOA"},
		{Name: "example.com.", Type: "MX"},
	}

	sortRRSets(records)

	if records[0].Type != "SOA" {
		t.Errorf("expected SOA first, got %s", records[0].Type)
	}
	if records[1].Type != "NS" {
		t.Errorf("expected NS second, got %s", records[1].Type)
	}
}

func TestFindSOATTY(t *testing.T) {
	records := []models.RRSet{
		{Name: "example.com.", Type: "SOA", TTL: 7200},
		{Name: "example.com.", Type: "NS", TTL: 3600},
	}

	ttl := findSOATTY(records)
	if ttl != 7200 {
		t.Errorf("expected 7200, got %d", ttl)
	}
}

func TestFindSOATTY_Default(t *testing.T) {
	records := []models.RRSet{
		{Name: "example.com.", Type: "NS", TTL: 3600},
	}

	ttl := findSOATTY(records)
	if ttl != 3600 {
		t.Errorf("expected default 3600, got %d", ttl)
	}
}

func TestFormatRecordContent(t *testing.T) {
	tests := []struct {
		rtype, content string
		priority       int
		expected       string
	}{
		{"MX", "mail.example.com.", 10, "10 mail.example.com."},
		{"TXT", "unquoted", 0, `"unquoted"`},
		{"TXT", `"already quoted"`, 0, `"already quoted"`},
		{"A", "192.0.2.1", 0, "192.0.2.1"},
		{"CNAME", "example.com.", 0, "example.com."},
	}

	for _, tc := range tests {
		t.Run(tc.rtype, func(t *testing.T) {
			result := formatRecordContent(tc.rtype, tc.content, tc.priority)
			if result != tc.expected {
				t.Errorf("formatRecordContent(%q, %q, %d) = %q, want %q", tc.rtype, tc.content, tc.priority, result, tc.expected)
			}
		})
	}
}

func importPDNS() testutil.PDNSHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/zones/") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/zones/") {
			w.Write([]byte(`{"id":"example.com.","name":"example.com.","kind":"Native","serial":2024010100}`))
			return
		}
		w.Write([]byte(`[]`))
	}
}

func TestImportZone_BIND(t *testing.T) {
	h, srv := newTestHandlerWithPDNS(t, importPDNS())
	defer srv.Close()

	bindContent := `$ORIGIN example.com.
$TTL 3600
@ IN SOA ns1.example.com. hostmaster.example.com. 2024010100 3600 900 1209600 3600
@ IN NS ns1.example.com.
www IN A 192.0.2.1
@ IN MX 10 mail.example.com.`

	body := fmt.Sprintf("--boundary\r\nContent-Disposition: form-data; name=\"zonefile\"; filename=\"test.zone\"\r\nContent-Type: text/plain\r\n\r\n%s\r\n--boundary--\r\n", bindContent)
	r := httptest.NewRequest(http.MethodPost, "/zones/example.com./import", strings.NewReader(body))
	r.SetPathValue("zone_id", "example.com.")
	r.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	r = withUserContext(r, &models.User{ID: 1, Username: "test", Role: "admin"})

	w := httptest.NewRecorder()
	h.ImportZone(w, r)
	assertImportRedirect(t, w)
}

func TestImportZone_CSV(t *testing.T) {
	h, srv := newTestHandlerWithPDNS(t, importPDNS())
	defer srv.Close()

	csvContent := "name,type,content,ttl,priority,disabled\n@,SOA,\"ns1.example.com. hostmaster.example.com. 2024010100 3600 900 1209600 3600\",3600,0,false\n@,NS,ns1.example.com.,3600,0,false\nwww,A,192.0.2.1,3600,0,false"

	body := fmt.Sprintf("--boundary\r\nContent-Disposition: form-data; name=\"zonefile\"; filename=\"test.csv\"\r\nContent-Type: text/csv\r\n\r\n%s\r\n--boundary--\r\n", csvContent)
	r := httptest.NewRequest(http.MethodPost, "/zones/example.com./import", strings.NewReader(body))
	r.SetPathValue("zone_id", "example.com.")
	r.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	r = withUserContext(r, &models.User{ID: 1, Username: "test", Role: "admin"})

	w := httptest.NewRecorder()
	h.ImportZone(w, r)
	assertImportRedirect(t, w)
}

func TestImportZone_NoFile(t *testing.T) {
	h, srv := newTestHandlerWithPDNS(t, importPDNS())
	defer srv.Close()

	r := httptest.NewRequest(http.MethodPost, "/zones/example.com./import", nil)
	r.SetPathValue("zone_id", "example.com.")
	r.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	r = withUserContext(r, &models.User{ID: 1, Username: "test", Role: "admin"})

	w := httptest.NewRecorder()
	h.ImportZone(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func assertImportRedirect(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestParseBindZone(t *testing.T) {
	data := []byte(`$ORIGIN example.com.
$TTL 3600
@ IN SOA ns1.example.com. hostmaster.example.com. 2024010100 3600 900 1209600 3600
@ IN NS ns1.example.com.
www IN A 192.0.2.1
@ IN MX 10 mail.example.com.`)

	rrsets, err := parseBindZone(data, "example.com.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rrsets) != 4 {
		t.Fatalf("expected 4 rrsets, got %d", len(rrsets))
	}

	types := map[string]bool{}
	for _, rr := range rrsets {
		types[rr.Type] = true
	}
	if !types["SOA"] || !types["NS"] || !types["A"] || !types["MX"] {
		t.Errorf("missing expected types: %v", types)
	}
}

func TestParseBindZone_Parens(t *testing.T) {
	data := []byte(`$ORIGIN example.com.
@ IN SOA ns1.example.com. hostmaster.example.com. (
    2024010100
    3600
    900
    1209600
    3600 )
@ IN NS ns1.example.com.`)

	rrsets, err := parseBindZone(data, "example.com.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rrsets) != 2 {
		t.Fatalf("expected 2 rrsets, got %d", len(rrsets))
	}
	if rrsets[0].Type != "SOA" {
		t.Errorf("expected SOA, got %s", rrsets[0].Type)
	}
	if len(rrsets[0].Records) != 1 {
		t.Errorf("expected 1 SOA record, got %d", len(rrsets[0].Records))
	}
}

func TestParseBindZone_IncludeDirective(t *testing.T) {
	data := []byte(`$ORIGIN example.com.
$INCLUDE /etc/bind/zones/other.zone
@ IN NS ns1.example.com.`)

	rrsets, err := parseBindZone(data, "example.com.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rrsets) != 1 {
		t.Fatalf("expected 1 rrset, got %d", len(rrsets))
	}
}

func TestParseCSVZone(t *testing.T) {
	input := `name,type,content,ttl,priority,disabled
@,SOA,"ns1.example.com. hostmaster.example.com. 2024010100 3600 900 1209600 3600",3600,0,false
example.com.,NS,ns1.example.com.,3600,0,false
www.example.com.,A,192.0.2.1,3600,0,false`

	rrsets, err := parseCSVZone(csv.NewReader(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rrsets) != 3 {
		t.Fatalf("expected 3 rrsets, got %d", len(rrsets))
	}
}

func TestParseCSVZone_NoData(t *testing.T) {
	input := `name,type,content,ttl,priority,disabled`
	rrsets, err := parseCSVZone(csv.NewReader(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rrsets) != 0 {
		t.Errorf("expected 0 rrsets, got %d", len(rrsets))
	}
}

func TestDetectFormat(t *testing.T) {
	if format := detectFormat("zone.csv"); format != "csv" {
		t.Errorf("expected csv, got %s", format)
	}
	if format := detectFormat("zone.zone"); format != "bind" {
		t.Errorf("expected bind, got %s", format)
	}
	if format := detectFormat("zone"); format != "bind" {
		t.Errorf("expected bind for unknown, got %s", format)
	}
}

func TestResolveBindName(t *testing.T) {
	origin := "example.com."

	tests := []struct {
		name, expected string
	}{
		{"@", "example.com."},
		{"www.example.com.", "www.example.com."},
		{"www", "www.example.com."},
	}

	for _, tc := range tests {
		result := resolveBindName(tc.name, origin)
		if result != tc.expected {
			t.Errorf("resolveBindName(%q, %q) = %q, want %q", tc.name, origin, result, tc.expected)
		}
	}
}

func TestExtractPriority(t *testing.T) {
	if p := extractPriority("MX", "10 mail.example.com."); p != 10 {
		t.Errorf("expected 10, got %d", p)
	}
	if p := extractPriority("SRV", "10 5 5060 sip.example.com."); p != 10 {
		t.Errorf("expected 10, got %d", p)
	}
	if p := extractPriority("A", "192.0.2.1"); p != 0 {
		t.Errorf("expected 0 for A, got %d", p)
	}
}

func TestGetCSVField(t *testing.T) {
	headers := map[string]int{"name": 0, "type": 1, "content": 2}
	row := []string{"example.com.", "A", "192.0.2.1"}

	if v := getCSVField(row, headers, "name"); v != "example.com." {
		t.Errorf("expected example.com., got %s", v)
	}
	if v := getCSVField(row, headers, "nonexistent"); v != "" {
		t.Errorf("expected empty, got %s", v)
	}
}
