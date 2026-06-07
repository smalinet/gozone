# ROADMAP - GoZone

Remaining tasks to improve the security, quality, and performance of GoZone.

## Performance

### HTTP Optimizations

- [x] **Configure connection pooling for PowerDNS client** (`internal/pdns/client.go`)
  ```go
  Transport: &http.Transport{
      MaxIdleConns:        100,
      MaxIdleConnsPerHost: 10,
      IdleConnTimeout:     90 * time.Second,
      DisableKeepAlives:   false,
  }
  ```

### Database Optimizations

- [x] **Cache frequently accessed data**
  - Zone cache (TTL 1 min) — `internal/pdns/cached.go`
  - PowerDNS statistics cache (TTL 30 sec) — `internal/pdns/cached.go`
  - TSIG keys cache (TTL 5 min) — `internal/pdns/cached.go`
  - Server info cache (TTL 5 min) — `internal/pdns/cached.go`
  - Invalidate zone/stat caches on CreateZone/DeleteZone
  - Invalidate TSIG cache on CreateTSIGKey/UpdateTSIGKey/DeleteTSIGKey
  - Generic TTL cache in `internal/cache/cache.go`
  - Wired in `cmd/gozone/main.go` via `pdns.NewCachedClient()`

## DNSSEC

- [x] **Zone DNSSEC management**
  - Enable/disable DNSSEC per zone via cryptokey creation/deletion
  - View DNSSEC keys on zone page with KSK/ZSK badge, algorithm, bits, DS records
  - Display active and inactive keys with activate/deactivate toggle
  - Create/delete keys via PowerDNS API in `internal/pdns/client.go`

- [x] **Key operations**
  - Create new KSK and ZSK keys per zone — `CreateCryptokey` handler
  - Delete deactivated keys — `DeleteCryptokey` handler
  - Toggle key active/inactive — `ToggleCryptokey` handler
  - Export DS records displayed inline on zone page

- [x] **DNSSEC algorithms**
  - Support for common algorithms (rsasha256, rsasha512, ecdsa256, ecdsa384, ed25519, ed448)
  - Display algorithm details in key table
  - Algorithm selection dropdown when creating new keys — `DNSSECAlgorithms()` helper
  - ECDSAP256SHA256 (ecdsa256) pre-selected as default

- [x] **Zone metadata for DNSSEC**
  - `NSEC3PARAM` metadata available through existing metadata management UI
  - `PRESIGNED` flag available through metadata UI
  - `PUBLISH-CDS` / `PUBLISH-CDNSKEY` available through metadata UI
  - Metadata kinds list includes all 14 DNSSEC-related kinds

## Export / Import

- [x] **Zone export in BIND format**
  - Generate RFC 1035-compliant zone file from PowerDNS records — `ExportZone` handler in `internal/handlers/export.go`
  - Include `$ORIGIN`, `$TTL`, and SOA record header
  - Support all record types (A, AAAA, CNAME, MX, NS, PTR, SOA, SRV, TXT, etc.)
  - Download as `.zone` file via button on zone view page
  - Server-side generation in Go (no external tools)
  - `GET /zones/{zone_id}/export?format=bind`

- [x] **Zone export in CSV format**
  - Columns: name, type, content, ttl, priority, disabled
  - `GET /zones/{zone_id}/export?format=csv`
  - `Content-Type: text/csv` with Content-Disposition attachment

- [x] **Import zone from BIND zone file**
  - Parse RFC 1035 zone file with `$ORIGIN`, `$TTL`, `$INCLUDE` directives via `parseBindZone()` in `internal/handlers/import.go`
  - Handle comments (`;` to EOL), parentheses for multi-line records, relative/absolute names, `@` origin expansion
  - Quoted TXT records preserved
  - File upload form on zone view page
  - Batch-create parsed records via `CreateRecords` PDNS API
  - `POST /zones/{zone_id}/import`

- [x] **Import zone from CSV file**
  - Parse CSV with header row: name, type, content, ttl, priority, disabled
  - Automatic trailing dot normalization on record names
  - Group records by (name, type) into RRSets before creating
  - `POST /zones/{zone_id}/import`

- [x] **Export / Import API endpoints**
  - Export available to all authenticated users with zone access
  - Import restricted to users with zone access (group authorization)
  - Respect zone-level authorization via `CheckZoneAccess` middleware
  - Set appropriate `Content-Type` headers (`text/plain` for BIND, `text/csv` for CSV)
  - File upload limited to 10MB with `ParseMultipartForm`

## Zone Templates

- [x] **Define reusable zone templates**
  - Pre-populated record sets for common zone types (standard, mail, web, redirect)
  - Custom templates with user-defined records
  - Template variables (e.g. `{{IP}}`, `{{MX_HOST}}`, `{{TTL}}`, `{{ZONE}}`)
  - Template storage in local database tables `zone_templates` + `zone_template_records`
  - Built-in templates seeded on startup: `standard`, `mail`, `web`, `redirect`

- [x] **Template management UI**
  - CRUD pages for templates (list, create, edit, delete) at `/templates`
  - Record editor within template edit page (add/delete records)
  - Template variables reference table on edit page
  - Template list shows built-in vs custom templates

- [x] **Apply template on zone creation**
  - Select template from dropdown on zone create page
  - Replace variables with user-provided values (IP, IP6, MX_HOST, etc.)
  - Batch-create all template records via PowerDNS API after zone creation
  - Auto-populate `{{ZONE}}` variable with the created zone name

- [x] **Apply template to existing zone**
  - "Apply Template" dropdown on zone view page
  - Same variable substitution as creation flow
  - Records are added to the existing zone via `POST /zones/{zone_id}/apply-template`
  - Route protected by zone-level group authorization

- [x] **Built-in templates**
  - `standard` — SOA + NS records
  - `mail` — SOA + NS + MX + SPF + DKIM placeholders
  - `web` — SOA + NS + A/AAAA + CNAME www
  - `redirect` — SOA + NS + A + URL redirect record

## Code Quality

### Cleanup

- [ ] **Add TODO/FIXME comments** for known limitations

### Monitoring and Observability

- [ ] **Add Prometheus metrics**
  - Request count per endpoint
  - Request latency
  - Errors by type
  - Zone/record/user counts
  - PowerDNS response time

- [ ] **Add distributed tracing**
  - Use `go.opentelemetry.io/otel`
  - Trace PowerDNS calls
  - Trace SQL queries
  - Export to Jaeger or Zipkin

## Performance Targets

- [ ] Average response time < 100ms for API endpoints
- [ ] Average response time < 500ms for web pages
- [ ] Support 100 requests/second with SQLite
- [ ] Support 1000 requests/second with PostgreSQL (future)

---

## Notes

### Known SQLite Limitations
- No multi-writer support (hence `SetMaxOpenConns(1)`)
- No native replication
- No clustering
- Limited to ~100 writes/second under load
- Recommended only for development and small installations

### Useful References
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Checklist](https://github.com/guardrailsio/awesome-golang-security)
- [PowerDNS API Documentation](https://doc.powerdns.com/authoritative/http-api/)
- [RFC 1035 - Domain Names](https://tools.ietf.org/html/rfc1035)
- [PowerDNS DNSSEC Guide](https://doc.powerdns.com/authoritative/dnssec/)
