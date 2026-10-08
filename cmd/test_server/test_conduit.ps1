# Conduit Full Integration Test Script
# Tests: schema mutation, CRUD, multi-format I/O, filters, ordering, pagination, errors, auth

$ErrorActionPreference = 'Stop'

function Write-Section($title) {
    Write-Host ''
    Write-Host "=== $title ===" -ForegroundColor Cyan
}

function Call-Conduit($method, $url, $body = $null, $headers = @{}) {
    $params = @{
        Method  = $method
        Uri     = $url
        Headers = $headers
    }
    if ($body) {
        $params.Body = $body
        $params.ContentType = 'application/json'
    }
    try {
        return Invoke-RestMethod @params
    }
    catch {
        Write-Host "ERROR calling $method $url" -ForegroundColor Red
        Write-Host $_.Exception.Message
        return $null
    }
}

# ------------------------------------------------------------------------------
# CONFIG
# ------------------------------------------------------------------------------

$base = 'http://localhost:8080/v1'
Write-Section 'Starting Conduit Test Suite'

# ------------------------------------------------------------------------------
# 1. SCHEMA CREATION
# ------------------------------------------------------------------------------

Write-Section 'Creating test table: players'

$schema = @'
{
  "columns": [
    { "name": "id", "type": "INTEGER", "pk": true, "autoincrement": true },
    { "name": "name", "type": "TEXT" },
    { "name": "score", "type": "INTEGER" },
    { "name": "active", "type": "BOOLEAN", "default": true }
  ]
}
'@

Call-Conduit 'POST' "$base/schema/players" $schema

# ------------------------------------------------------------------------------
# 2. INSERT ROWS
# ------------------------------------------------------------------------------

Write-Section 'Inserting rows'

Call-Conduit 'POST' "$base/players" '{"name":"Alice","score":98,"active":true}'
Call-Conduit 'POST' "$base/players" '{"name":"Bob","score":72,"active":false}'
Call-Conduit 'POST' "$base/players" '{"name":"Charlie","score":88,"active":true}'

# ------------------------------------------------------------------------------
# 3. LIST ENDPOINT (JSON)
# ------------------------------------------------------------------------------

Write-Section 'List players (JSON)'
Call-Conduit 'GET' "$base/players?format=json"

# ------------------------------------------------------------------------------
# 4. LIST ENDPOINT WITH WHERE FILTER
# ------------------------------------------------------------------------------

Write-Section 'List players WHERE active=true'
Call-Conduit 'GET' "$base/players?where=active=true"

Write-Section 'List players WHERE score>80'
Call-Conduit 'GET' "$base/players?where=score>80"

# ------------------------------------------------------------------------------
# 5. ORDER BY
# ------------------------------------------------------------------------------

Write-Section 'List players ORDER BY score DESC'
Call-Conduit 'GET' "$base/players?orderBy=score desc"

Write-Section 'List players ORDER BY name ASC'
Call-Conduit 'GET' "$base/players?orderBy=name asc"

# ------------------------------------------------------------------------------
# 6. PAGINATION
# ------------------------------------------------------------------------------

Write-Section 'Pagination test (limit=1 offset=1)'
Call-Conduit 'GET' "$base/players?limit=1&offset=1"

# ------------------------------------------------------------------------------
# 7. GET SINGLE ROW
# ------------------------------------------------------------------------------

Write-Section 'Get single row (id=1)'
Call-Conduit 'GET' "$base/players/1"

# ------------------------------------------------------------------------------
# 8. UPDATE ROW (PUT)
# ------------------------------------------------------------------------------

Write-Section 'Updating row id=2'
Call-Conduit 'PUT' "$base/players/2" '{"name":"Bobby","score":75}'

# ------------------------------------------------------------------------------
# 9. PATCH ROW
# ------------------------------------------------------------------------------

Write-Section 'Patching row id=3'
Call-Conduit 'PATCH' "$base/players/3" '{"score":90}'

# ------------------------------------------------------------------------------
# 10. DELETE ROW
# ------------------------------------------------------------------------------

Write-Section 'Deleting row id=2'
Call-Conduit 'DELETE' "$base/players/2"

# ------------------------------------------------------------------------------
# 11. MULTI-FORMAT OUTPUT TESTS
# ------------------------------------------------------------------------------
Write-Section 'JSON Output'
Call-Conduit 'GET' "$base/players?format=json"

Write-Section 'XML Output'
Call-Conduit 'GET' "$base/players?format=xml"

Write-Section 'YAML Output'
Call-Conduit 'GET' "$base/players?format=yaml"

Write-Section 'TOML Output'
Call-Conduit 'GET' "$base/players?format=toml"

Write-Section 'CSV Output'
Call-Conduit 'GET' "$base/players?format=csv"

Write-Section 'NDJSON Output'
Call-Conduit 'GET' "$base/players?format=ndjson"

# ------------------------------------------------------------------------------
# 12. SCHEMA EXPORT (ALL FORMATS)
# ------------------------------------------------------------------------------

Write-Section 'Schema Export (JSON)'
Call-Conduit 'GET' "$base/schema/players?format=json"

Write-Section 'Schema Export (XML)'
Call-Conduit 'GET' "$base/schema/players?format=xml"

Write-Section 'Schema Export (YAML)'
Call-Conduit 'GET' "$base/schema/players?format=yaml"

Write-Section 'Schema Export (CSV)'
Call-Conduit 'GET' "$base/schema/players?format=csv"

Write-Section 'Schema Export (TOML)'
Call-Conduit 'GET' "$base/schema/players?format=toml"

Write-Section 'Schema Export (NDJSON)'
Call-Conduit 'GET' "$base/schema/players?format=ndjson"

# ------------------------------------------------------------------------------
# 13. ERROR TESTS
# ------------------------------------------------------------------------------

Write-Section 'Error: GET missing record'
Call-Conduit 'GET' "$base/players/999?format=json"

Write-Section 'Error: Invalid WHERE syntax'
Call-Conduit 'GET' "$base/players?where=score>>80"

Write-Section 'Error: Invalid ORDER BY'
Call-Conduit 'GET' "$base/players?orderBy=unknown desc"

# ------------------------------------------------------------------------------
# 14. DELETE TABLE
# ------------------------------------------------------------------------------

Write-Section 'Deleting table players'
Call-Conduit 'DELETE' "$base/schema/players"

Write-Section 'Final schema list'
Call-Conduit 'GET' "$base/schema?format=json"

Write-Host ''
Write-Host '=== Conduit Test Suite Complete ===' -ForegroundColor Green
