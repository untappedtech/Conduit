# ==============================================================================
# Conduit Live Server Comprehensive Integration Test Suite
# Tests:
#   - OpenAPI 3.0 specification & Scalar docs discovery
#   - Schema mutation in ALL formats (JSON, YAML, TOML, XML, CSV, CBOR)
#   - Schema introspection and table listing in ALL formats
#   - Full CRUD operations in EACH format (JSON, YAML, TOML, XML, CSV, CBOR, NDJSON)
#   - Advanced querying: WHERE DSL (operators, grouping), ORDER BY, pagination
#   - Standard multi-format error envelopes
#   - Schema teardown and cleanup
# ==============================================================================

[CmdletBinding()]
param (
    [string]$BaseUrl = 'http://localhost:8080',
    [string]$ApiKey = '',
    [switch]$VerboseOutput
)

$ErrorActionPreference = 'Continue'

# ------------------------------------------------------------------------------
# Test Harness State & Helpers
# ------------------------------------------------------------------------------

$script:TotalCount = 0
$script:PassCount = 0
$script:FailCount = 0

function Write-Banner($title) {
    Write-Host ''
    Write-Host '======================================================================' -ForegroundColor Cyan
    Write-Host "  $title" -ForegroundColor Cyan
    Write-Host '======================================================================' -ForegroundColor Cyan
}

function Write-Section($title) {
    Write-Host ''
    Write-Host "--- $title ---" -ForegroundColor Yellow
}

function Invoke-ConduitRequest {
    param (
        [string]$Method,
        [string]$Url,
        $Body = $null,
        [string]$ContentType = $null,
        [hashtable]$Headers = @{}
    )

    $reqHeaders = @{}
    foreach ($k in $Headers.Keys) {
        $reqHeaders[$k] = $Headers[$k]
    }
    if ($ApiKey) {
        $reqHeaders['Authorization'] = "Bearer $ApiKey"
    }

    $params = @{
        Method             = $Method
        Uri                = $Url
        Headers            = $reqHeaders
        SkipHttpErrorCheck = $true
        TimeoutSec         = 15
    }

    if ($Body -is [byte[]]) {
        $params.Body = $Body
    }
    elseif ($null -ne $Body -and $Body -ne '') {
        $params.Body = [System.Text.Encoding]::UTF8.GetBytes($Body)
    }

    if ($ContentType) {
        $params.ContentType = $ContentType
    }

    try {
        $resp = Invoke-WebRequest @params
        $rawBytes = $null
        $contentStr = ''

        if ($resp.Content -is [byte[]]) {
            $rawBytes = $resp.Content
            $contentStr = [System.Text.Encoding]::UTF8.GetString($resp.Content)
        }
        elseif ($resp.Content -is [string]) {
            $contentStr = $resp.Content
            $rawBytes = [System.Text.Encoding]::UTF8.GetBytes($resp.Content)
        }
        elseif ($resp.RawContentStream) {
            $mem = [System.IO.MemoryStream]::new()
            $resp.RawContentStream.Position = 0
            $resp.RawContentStream.CopyTo($mem)
            $rawBytes = $mem.ToArray()
            $contentStr = [System.Text.Encoding]::UTF8.GetString($rawBytes)
        }

        return [PSCustomObject]@{
            StatusCode  = [int]$resp.StatusCode
            Content     = $contentStr
            RawBytes    = $rawBytes
            ContentType = $resp.Headers['Content-Type']
            Headers     = $resp.Headers
            Success     = $true
            Error       = $null
        }
    }
    catch {
        return [PSCustomObject]@{
            StatusCode  = 0
            Content     = ''
            RawBytes    = $null
            ContentType = ''
            Headers     = @{}
            Success     = $false
            Error       = $_.Exception.Message
        }
    }
}

function Assert-Test {
    param (
        [string]$TestName,
        $Response,
        [int[]]$ExpectedStatus = @(200),
        [string]$ExpectedContentType = '',
        [string[]]$ExpectedSubstrings = @(),
        [string[]]$UnexpectedSubstrings = @()
    )

    $script:TotalCount++

    if (-not $Response -or -not $Response.Success) {
        $script:FailCount++
        Write-Host "  [FAIL] $TestName - Network/Transport Error: $($Response.Error)" -ForegroundColor Red
        return
    }

    $statusMatch = $ExpectedStatus -contains $Response.StatusCode
    $typeMatch = $true
    if ($ExpectedContentType) {
        $actualType = [string]$Response.ContentType
        $typeMatch = $actualType.ToLower().Contains($ExpectedContentType.ToLower())
    }

    $missingSubstrings = @()
    foreach ($sub in $ExpectedSubstrings) {
        if (-not $Response.Content.Contains($sub)) {
            $missingSubstrings += $sub
        }
    }

    $foundUnexpected = @()
    foreach ($unsub in $UnexpectedSubstrings) {
        if ($Response.Content.Contains($unsub)) {
            $foundUnexpected += $unsub
        }
    }

    if ($statusMatch -and $typeMatch -and ($missingSubstrings.Count -eq 0) -and ($foundUnexpected.Count -eq 0)) {
        $script:PassCount++
        Write-Host "  [PASS] $TestName (Status: $($Response.StatusCode))" -ForegroundColor Green
        if ($VerboseOutput) {
            Write-Host "         Content snippet: $($Response.Content.Trim().Substring(0, [Math]::Min(120, $Response.Content.Trim().Length)))" -ForegroundColor DarkGray
        }
        return
    }

    $script:FailCount++
    Write-Host "  [FAIL] $TestName" -ForegroundColor Red
    if (-not $statusMatch) {
        Write-Host "         Expected Status: $($ExpectedStatus -join ', '), Got: $($Response.StatusCode)" -ForegroundColor Red
    }
    if (-not $typeMatch) {
        Write-Host "         Expected Content-Type: '$ExpectedContentType', Got: '$($Response.ContentType)'" -ForegroundColor Red
    }
    if ($missingSubstrings.Count -gt 0) {
        Write-Host "         Missing Substrings: $($missingSubstrings -join ', ')" -ForegroundColor Red
    }
    if ($foundUnexpected.Count -gt 0) {
        Write-Host "         Found Unexpected Substrings: $($foundUnexpected -join ', ')" -ForegroundColor Red
    }
    if ($VerboseOutput -and $Response.Content) {
        Write-Host "         Full Body:`n$($Response.Content)" -ForegroundColor DarkRed
    }
    return $false
}

# ------------------------------------------------------------------------------
# Connectivity Pre-Check
# ------------------------------------------------------------------------------

$base = "$BaseUrl/v1"
$root = $BaseUrl.TrimEnd('/')

Write-Banner 'Conduit Full Live Integration Test Suite'
Write-Host "Target Server: $BaseUrl" -ForegroundColor Cyan
Write-Host "API Base:      $base" -ForegroundColor Cyan

$health = Invoke-ConduitRequest -Method 'GET' -Url "$base/schema"
if (-not $health.Success -or $health.StatusCode -eq 0) {
    Write-Host "`n[ERROR] Unable to connect to Conduit at $BaseUrl." -ForegroundColor Red
    Write-Host 'Please start the Conduit server before executing this test script:' -ForegroundColor Yellow
    Write-Host '  go run ./cmd/server' -ForegroundColor White
    Write-Host '  OR: .\conduit.exe' -ForegroundColor White
    exit 1
}

# Cleanup existing test tables if leftover from prior runs
foreach ($tbl in @('test_json', 'test_yaml', 'test_toml', 'test_xml', 'test_csv', 'test_cbor', 'test_ndjson')) {
    $null = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/schema/$tbl"
}

# ==============================================================================
# SECTION 1: OpenAPI & Interactive Documentation Endpoints
# ==============================================================================
Write-Section '1. OpenAPI 3.0 & Interactive Documentation Discovery'

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/openapi.json"
Assert-Test -TestName "GET $base/openapi.json returns valid spec" `
    -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/json' `
    -ExpectedSubstrings @('"openapi": "3.0.3"', '"title": "Conduit API"', '"/schema"')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$root/openapi.json"
Assert-Test -TestName "GET $root/openapi.json (root alias) returns valid spec" `
    -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/json' `
    -ExpectedSubstrings @('"openapi": "3.0.3"')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/docs"
Assert-Test -TestName "GET $base/docs returns Scalar UI HTML" `
    -Response $r -ExpectedStatus 200 -ExpectedContentType 'text/html' `
    -ExpectedSubstrings @('@scalar/api-reference', '<!doctype html>')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$root/docs"
Assert-Test -TestName "GET $root/docs (root alias) returns Scalar UI HTML" `
    -Response $r -ExpectedStatus 200 -ExpectedContentType 'text/html' `
    -ExpectedSubstrings @('@scalar/api-reference')

# ==============================================================================
# SECTION 2: Schema Mutation - Creating Tables in ALL Formats
# ==============================================================================
Write-Section '2. Schema Mutation Endpoints (POST /v1/schema/<table> in ALL Formats)'

# 2.1 JSON Schema
$schemaJSON = @'
{
  "columns": [
    { "name": "id", "type": "INTEGER", "pk": true, "autoincrement": true },
    { "name": "name", "type": "TEXT", "nullable": false },
    { "name": "score", "type": "INTEGER", "nullable": true },
    { "name": "active", "type": "BOOLEAN", "default": true }
  ]
}
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/schema/test_json" -Body $schemaJSON -ContentType 'application/json'
Assert-Test -TestName 'Create schema test_json (JSON input)' `
    -Response $r -ExpectedStatus @(200, 201) -ExpectedSubstrings @('"name": "id"', '"name": "name"')

# 2.2 YAML Schema
$schemaYAML = @'
columns:
  - name: id
    type: INTEGER
    pk: true
    autoincrement: true
  - name: name
    type: TEXT
    nullable: false
  - name: score
    type: INTEGER
  - name: active
    type: BOOLEAN
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/schema/test_yaml" -Body $schemaYAML -ContentType 'application/x-yaml'
Assert-Test -TestName 'Create schema test_yaml (YAML input)' `
    -Response $r -ExpectedStatus @(200, 201) -ExpectedSubstrings @('name: id', 'name: name')

# 2.3 TOML Schema
$schemaTOML = @'
[[columns]]
name = "id"
type = "INTEGER"
pk = true
autoincrement = true

[[columns]]
name = "name"
type = "TEXT"

[[columns]]
name = "score"
type = "INTEGER"

[[columns]]
name = "active"
type = "BOOLEAN"
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/schema/test_toml" -Body $schemaTOML -ContentType 'application/toml'
Assert-Test -TestName 'Create schema test_toml (TOML input)' `
    -Response $r -ExpectedStatus @(200, 201) -ExpectedSubstrings @('name = "id"', 'name = "name"')

# 2.4 XML Schema
$schemaXML = @'
<schema>
  <column>
    <name>id</name>
    <type>INTEGER</type>
    <pk>true</pk>
    <autoincrement>true</autoincrement>
  </column>
  <column>
    <name>name</name>
    <type>TEXT</type>
  </column>
  <column>
    <name>score</name>
    <type>INTEGER</type>
  </column>
  <column>
    <name>active</name>
    <type>BOOLEAN</type>
  </column>
</schema>
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/schema/test_xml" -Body $schemaXML -ContentType 'application/xml'
Assert-Test -TestName 'Create schema test_xml (XML input)' `
    -Response $r -ExpectedStatus @(200, 201) -ExpectedSubstrings @('<name>id</name>', '<name>name</name>')

# 2.5 CSV Schema
$schemaCSV = @'
name,type,pk,autoincrement
id,INTEGER,true,true
name,TEXT,false,false
score,INTEGER,false,false
active,BOOLEAN,false,false
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/schema/test_csv" -Body $schemaCSV -ContentType 'text/csv'
Assert-Test -TestName 'Create schema test_csv (CSV input)' `
    -Response $r -ExpectedStatus @(200, 201) -ExpectedSubstrings @('id,INTEGER', 'name,TEXT')

# 2.6 CBOR Schema (binary)
$cborSchemaHex = 'a167636f6c756d6e7383a462706bf56d6175746f696e6372656d656e74f5646e616d65626964647479706567494e5445474552a2646e616d65646e616d6564747970656454455854a2646e616d656573636f7265647479706567494e5445474552'
$cborSchemaBytes = [System.Convert]::FromHexString($cborSchemaHex)
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/schema/test_cbor" -Body $cborSchemaBytes -ContentType 'application/cbor'
Assert-Test -TestName 'Create schema test_cbor (CBOR input)' `
    -Response $r -ExpectedStatus @(200, 201) -ExpectedContentType 'application/cbor'

# ==============================================================================
# SECTION 3: Schema Inspection in ALL Output Formats
# ==============================================================================
Write-Section '3. Schema Export & Inspection (All 7 Output Formats)'

# 3.1 Inspect table schema
$formats = @(
    @{ Fmt = 'json'; CT = 'application/json'; Snippet = '"columns": [' },
    @{ Fmt = 'yaml'; CT = 'yaml'; Snippet = 'columns:' },
    @{ Fmt = 'toml'; CT = 'toml'; Snippet = '[[columns]]' },
    @{ Fmt = 'xml'; CT = 'xml'; Snippet = '<columns>' },
    @{ Fmt = 'csv'; CT = 'text/csv'; Snippet = 'name,type' },
    @{ Fmt = 'ndjson'; CT = 'ndjson'; Snippet = '"name":' },
    @{ Fmt = 'cbor'; CT = 'application/cbor'; Snippet = '' }
)

foreach ($f in $formats) {
    $r = Invoke-ConduitRequest -Method 'GET' -Url "$base/schema/test_json?format=$($f.Fmt)"
    $expectedSnips = if ($f.Snippet) { @($f.Snippet) } else { @() }
    Assert-Test -TestName "Export schema test_json ($($f.Fmt))" `
        -Response $r -ExpectedStatus 200 -ExpectedContentType $f.CT -ExpectedSubstrings $expectedSnips
}

# 3.2 List all tables in all formats
foreach ($f in $formats) {
    $r = Invoke-ConduitRequest -Method 'GET' -Url "$base/schema?format=$($f.Fmt)"
    $expectedSnips = if ($f.Fmt -ne 'cbor') { @('test_json') } else { @() }
    Assert-Test -TestName "List all tables ($($f.Fmt))" `
        -Response $r -ExpectedStatus 200 -ExpectedContentType $f.CT -ExpectedSubstrings $expectedSnips
}

# ==============================================================================
# SECTION 4: Full CRUD Operations in EACH Format
# ==============================================================================
Write-Section '4. CRUD Operations in EACH Format'

# 4.1 JSON CRUD (Table: test_json)
Write-Host "`n  -> Testing JSON CRUD" -ForegroundColor Magenta
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_json" -Body '{"name":"Alice","score":95,"active":true}' -ContentType 'application/json'
Assert-Test -TestName 'JSON CRUD: POST record 1' -Response $r -ExpectedStatus 201 -ExpectedContentType 'application/json' -ExpectedSubstrings @('"name": "Alice"')

$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_json" -Body '{"name":"Bob","score":80,"active":false}' -ContentType 'application/json'
Assert-Test -TestName 'JSON CRUD: POST record 2' -Response $r -ExpectedStatus 201

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?format=json"
Assert-Test -TestName 'JSON CRUD: GET collection' -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/json' -ExpectedSubstrings @('Alice', 'Bob')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json/1?format=json"
Assert-Test -TestName 'JSON CRUD: GET record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/json' -ExpectedSubstrings @('"name": "Alice"')

$r = Invoke-ConduitRequest -Method 'PUT' -Url "$base/test_json/1" -Body '{"name":"Alice M","score":99,"active":true}' -ContentType 'application/json'
Assert-Test -TestName 'JSON CRUD: PUT record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('"name": "Alice M"', '"score": 99')

$r = Invoke-ConduitRequest -Method 'PATCH' -Url "$base/test_json/1" -Body '{"score":100}' -ContentType 'application/json'
Assert-Test -TestName 'JSON CRUD: PATCH record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('"score": 100')

$r = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/test_json/2"
Assert-Test -TestName 'JSON CRUD: DELETE record 2' -Response $r -ExpectedStatus 204

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json/2?format=json"
Assert-Test -TestName 'JSON CRUD: Verify deleted record 2 is 404' -Response $r -ExpectedStatus 404

# 4.2 YAML CRUD (Table: test_yaml)
Write-Host "`n  -> Testing YAML CRUD" -ForegroundColor Magenta
$yamlInsert = @'
name: Charlie
score: 85
active: true
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_yaml" -Body $yamlInsert -ContentType 'application/x-yaml'
Assert-Test -TestName 'YAML CRUD: POST record' -Response $r -ExpectedStatus 201 -ExpectedContentType 'yaml' -ExpectedSubstrings @('name: Charlie')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_yaml?format=yaml"
Assert-Test -TestName 'YAML CRUD: GET collection' -Response $r -ExpectedStatus 200 -ExpectedContentType 'yaml' -ExpectedSubstrings @('test_yaml:', 'name: Charlie')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_yaml/1?format=yaml"
Assert-Test -TestName 'YAML CRUD: GET record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'yaml' -ExpectedSubstrings @('name: Charlie')

$yamlUpdate = @'
name: Charlie M
score: 88
active: true
'@
$r = Invoke-ConduitRequest -Method 'PUT' -Url "$base/test_yaml/1" -Body $yamlUpdate -ContentType 'application/x-yaml'
Assert-Test -TestName 'YAML CRUD: PUT record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('name: Charlie M')

$yamlPatch = @'
score: 92
'@
$r = Invoke-ConduitRequest -Method 'PATCH' -Url "$base/test_yaml/1" -Body $yamlPatch -ContentType 'application/x-yaml'
Assert-Test -TestName 'YAML CRUD: PATCH record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('score: 92')

$r = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/test_yaml/1"
Assert-Test -TestName 'YAML CRUD: DELETE record 1' -Response $r -ExpectedStatus 204

# 4.3 TOML CRUD (Table: test_toml)
Write-Host "`n  -> Testing TOML CRUD" -ForegroundColor Magenta
$tomlInsert = @'
name = "Dave"
score = 75
active = true
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_toml" -Body $tomlInsert -ContentType 'application/toml'
Assert-Test -TestName 'TOML CRUD: POST record' -Response $r -ExpectedStatus 201 -ExpectedContentType 'toml' -ExpectedSubstrings @('name = "Dave"')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_toml?format=toml"
Assert-Test -TestName 'TOML CRUD: GET collection' -Response $r -ExpectedStatus 200 -ExpectedContentType 'toml' -ExpectedSubstrings @('[[test_toml]]', 'name = "Dave"')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_toml/1?format=toml"
Assert-Test -TestName 'TOML CRUD: GET record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'toml' -ExpectedSubstrings @('[test_toml]', 'name = "Dave"')

$tomlUpdate = @'
name = "Dave M"
score = 78
active = true
'@
$r = Invoke-ConduitRequest -Method 'PUT' -Url "$base/test_toml/1" -Body $tomlUpdate -ContentType 'application/toml'
Assert-Test -TestName 'TOML CRUD: PUT record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('name = "Dave M"')

$tomlPatch = @'
score = 82
'@
$r = Invoke-ConduitRequest -Method 'PATCH' -Url "$base/test_toml/1" -Body $tomlPatch -ContentType 'application/toml'
Assert-Test -TestName 'TOML CRUD: PATCH record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('score = 82')

$r = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/test_toml/1"
Assert-Test -TestName 'TOML CRUD: DELETE record 1' -Response $r -ExpectedStatus 204

# 4.4 XML CRUD (Table: test_xml)
Write-Host "`n  -> Testing XML CRUD" -ForegroundColor Magenta
$xmlInsert = @'
<item>
  <name>Eve</name>
  <score>65</score>
  <active>true</active>
</item>
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_xml" -Body $xmlInsert -ContentType 'application/xml'
Assert-Test -TestName 'XML CRUD: POST record' -Response $r -ExpectedStatus 201 -ExpectedContentType 'xml' -ExpectedSubstrings @('<name>Eve</name>')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_xml?format=xml"
Assert-Test -TestName 'XML CRUD: GET collection' -Response $r -ExpectedStatus 200 -ExpectedContentType 'xml' -ExpectedSubstrings @('<test_xml>', '<row>', '<name>Eve</name>')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_xml/1?format=xml"
Assert-Test -TestName 'XML CRUD: GET record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'xml' -ExpectedSubstrings @('<test_xml>', '<name>Eve</name>')

$xmlUpdate = @'
<item>
  <name>Eve M</name>
  <score>69</score>
  <active>true</active>
</item>
'@
$r = Invoke-ConduitRequest -Method 'PUT' -Url "$base/test_xml/1" -Body $xmlUpdate -ContentType 'application/xml'
Assert-Test -TestName 'XML CRUD: PUT record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('<name>Eve M</name>')

$xmlPatch = @'
<item>
  <score>72</score>
</item>
'@
$r = Invoke-ConduitRequest -Method 'PATCH' -Url "$base/test_xml/1" -Body $xmlPatch -ContentType 'application/xml'
Assert-Test -TestName 'XML CRUD: PATCH record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('<score>72</score>')

$r = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/test_xml/1"
Assert-Test -TestName 'XML CRUD: DELETE record 1' -Response $r -ExpectedStatus 204

# 4.5 CSV CRUD (Table: test_csv)
Write-Host "`n  -> Testing CSV CRUD" -ForegroundColor Magenta
$csvInsert = @'
name,score,active
Frank,55,true
'@
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_csv" -Body $csvInsert -ContentType 'text/csv'
Assert-Test -TestName 'CSV CRUD: POST record' -Response $r -ExpectedStatus 201 -ExpectedContentType 'text/csv' -ExpectedSubstrings @('Frank')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_csv?format=csv"
Assert-Test -TestName 'CSV CRUD: GET collection' -Response $r -ExpectedStatus 200 -ExpectedContentType 'text/csv' -ExpectedSubstrings @('id,name,score,active', 'Frank')

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_csv/1?format=csv"
Assert-Test -TestName 'CSV CRUD: GET record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'text/csv' -ExpectedSubstrings @('Frank')

$csvUpdate = @'
name,score,active
Frank M,59,true
'@
$r = Invoke-ConduitRequest -Method 'PUT' -Url "$base/test_csv/1" -Body $csvUpdate -ContentType 'text/csv'
Assert-Test -TestName 'CSV CRUD: PUT record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Frank M')

$csvPatch = @'
score
63
'@
$r = Invoke-ConduitRequest -Method 'PATCH' -Url "$base/test_csv/1" -Body $csvPatch -ContentType 'text/csv'
Assert-Test -TestName 'CSV CRUD: PATCH record 1' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('63')

$r = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/test_csv/1"
Assert-Test -TestName 'CSV CRUD: DELETE record 1' -Response $r -ExpectedStatus 204

# 4.6 CBOR CRUD (Table: test_cbor)
Write-Host "`n  -> Testing CBOR CRUD" -ForegroundColor Magenta
$cborItemBytes = [System.Convert]::FromHexString('a2646e616d65634574616573636f72651846')
$r = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_cbor" -Body $cborItemBytes -ContentType 'application/cbor'
Assert-Test -TestName 'CBOR CRUD: POST record' -Response $r -ExpectedStatus 201 -ExpectedContentType 'application/cbor'

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_cbor?format=cbor"
Assert-Test -TestName 'CBOR CRUD: GET collection' -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/cbor'

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_cbor/1?format=cbor"
Assert-Test -TestName 'CBOR CRUD: GET record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/cbor'

$cborUpdateBytes = [System.Convert]::FromHexString('a2646e616d656b45746120557064617465646573636f7265184b')
$r = Invoke-ConduitRequest -Method 'PUT' -Url "$base/test_cbor/1" -Body $cborUpdateBytes -ContentType 'application/cbor'
Assert-Test -TestName 'CBOR CRUD: PUT record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/cbor'

$cborPatchBytes = [System.Convert]::FromHexString('a16573636f72651850')
$r = Invoke-ConduitRequest -Method 'PATCH' -Url "$base/test_cbor/1" -Body $cborPatchBytes -ContentType 'application/cbor'
Assert-Test -TestName 'CBOR CRUD: PATCH record 1' -Response $r -ExpectedStatus 200 -ExpectedContentType 'application/cbor'

$r = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/test_cbor/1"
Assert-Test -TestName 'CBOR CRUD: DELETE record 1' -Response $r -ExpectedStatus 204

# 4.7 NDJSON Stream Output
Write-Host "`n  -> Testing NDJSON Output" -ForegroundColor Magenta
$null = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_json" -Body '{"name":"NDJSON_1","score":10}' -ContentType 'application/json'
$null = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_json" -Body '{"name":"NDJSON_2","score":20}' -ContentType 'application/json'

$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?format=ndjson"
Assert-Test -TestName 'NDJSON: GET collection as newline stream' `
    -Response $r -ExpectedStatus 200 -ExpectedContentType 'ndjson' -ExpectedSubstrings @('"name":"NDJSON_1"', '"name":"NDJSON_2"')

# ==============================================================================
# SECTION 5: Query Capabilities (Where DSL, Order, Pagination)
# ==============================================================================
Write-Section '5. Query Capabilities (WHERE DSL, ORDER BY, Pagination)'

# Insert seed dataset into test_json
$null = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_json" -Body '{"name":"Filter_High","score":95,"active":true}' -ContentType 'application/json'
$null = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_json" -Body '{"name":"Filter_Mid","score":50,"active":true}' -ContentType 'application/json'
$null = Invoke-ConduitRequest -Method 'POST' -Url "$base/test_json" -Body '{"name":"Filter_Low","score":15,"active":false}' -ContentType 'application/json'

# WHERE: Greater than
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?where=score > 60&format=json"
Assert-Test -TestName 'WHERE filter: score > 60' `
    -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_High') -UnexpectedSubstrings @('Filter_Low', 'Filter_Mid')

# WHERE: Equality and Boolean
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?where=active = false&format=json"
Assert-Test -TestName 'WHERE filter: active = false' `
    -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_Low') -UnexpectedSubstrings @('Filter_High')

# WHERE: LIKE pattern matching
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?where=name LIKE '%Mid%'&format=json"
Assert-Test -TestName "WHERE filter: LIKE '%Mid%'" `
    -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_Mid') -UnexpectedSubstrings @('Filter_High')

# WHERE: IN set membership
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?where=score IN (15, 95)&format=json"
Assert-Test -TestName 'WHERE filter: IN list' `
    -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_High', 'Filter_Low') -UnexpectedSubstrings @('Filter_Mid')

# WHERE: Compound expression with parentheses and AND/OR
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?where=(score >= 90 OR score <= 20) AND active = true&format=json"
Assert-Test -TestName 'WHERE filter: complex expression with precedence' `
    -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_High') -UnexpectedSubstrings @('Filter_Low')

# ORDER BY: Column Descending
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?order=score:desc&format=json"
Assert-Test -TestName 'ORDER filter: score:desc' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_High')

# ORDER BY: Column Ascending
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?order=score:asc&format=json"
Assert-Test -TestName 'ORDER filter: score:asc' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_Low')

# Pagination: limit & offset
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?limit=1&offset=0&format=json"
Assert-Test -TestName 'Pagination: limit=1 offset=0' -Response $r -ExpectedStatus 200

# Unlimited pagination mode: limit=0
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?limit=0&format=json"
Assert-Test -TestName 'Pagination: limit=0 (unlimited)' -Response $r -ExpectedStatus 200 -ExpectedSubstrings @('Filter_High', 'Filter_Mid', 'Filter_Low')

# ==============================================================================
# SECTION 6: Standard Error Envelopes Across All Formats
# ==============================================================================
Write-Section '6. Standard Error Envelopes in ALL Formats'

$errorFormats = @(
    @{ Fmt = 'json'; CT = 'application/json'; Snippet = '"title": "Not Found"' },
    @{ Fmt = 'yaml'; CT = 'yaml'; Snippet = 'title: "Not Found"' },
    @{ Fmt = 'toml'; CT = 'toml'; Snippet = 'title = "Not Found"' },
    @{ Fmt = 'xml'; CT = 'xml'; Snippet = '<title>Not Found</title>' },
    @{ Fmt = 'csv'; CT = 'text/csv'; Snippet = 'title,Not Found' },
    @{ Fmt = 'ndjson'; CT = 'ndjson'; Snippet = '"title":"Not Found"' },
    @{ Fmt = 'cbor'; CT = 'application/cbor'; Snippet = '' }
)

foreach ($ef in $errorFormats) {
    $r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json/999999?format=$($ef.Fmt)"
    $expectedSnips = if ($ef.Snippet) { @($ef.Snippet) } else { @() }
    Assert-Test -TestName "404 Error envelope ($($ef.Fmt))" `
        -Response $r -ExpectedStatus 404 -ExpectedContentType $ef.CT -ExpectedSubstrings $expectedSnips
}

# 400 Bad Request: Invalid sort column
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?order=nonexistent_col:asc"
Assert-Test -TestName '400 Error: Invalid ORDER BY column' -Response $r -ExpectedStatus 400 -ExpectedSubstrings @('Bad Request')

# 422 Unprocessable Entity: Malformed WHERE syntax
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/test_json?where=score >> 999"
Assert-Test -TestName '422 Error: Malformed WHERE expression syntax' -Response $r -ExpectedStatus 422 -ExpectedSubstrings @('Unprocessable Entity')

# 405 Method Not Allowed
$r = Invoke-ConduitRequest -Method 'PUT' -Url "$base/schema/test_json" -Body '{}' -ContentType 'application/json'
Assert-Test -TestName '405 Error: PUT /schema/<table> is Method Not Allowed' -Response $r -ExpectedStatus 405

# ==============================================================================
# SECTION 7: Schema Teardown (DELETE /v1/schema/<table>)
# ==============================================================================
Write-Section '7. Schema Teardown & Dynamic Invalidation'

$testTables = @('test_json', 'test_yaml', 'test_toml', 'test_xml', 'test_csv', 'test_cbor')
foreach ($tbl in $testTables) {
    $r = Invoke-ConduitRequest -Method 'DELETE' -Url "$base/schema/$tbl"
    Assert-Test -TestName "DELETE /schema/$tbl" -Response $r -ExpectedStatus 204
}

# Verify table list after drops
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/schema?format=json"
Assert-Test -TestName 'Verify test tables removed from /schema list' `
    -Response $r -ExpectedStatus 200 -UnexpectedSubstrings @('"test_json"', '"test_yaml"', '"test_toml"')

# Verify OpenAPI spec no longer includes dropped tables
$r = Invoke-ConduitRequest -Method 'GET' -Url "$base/openapi.json"
Assert-Test -TestName 'Verify OpenAPI spec invalidated and test tables removed' `
    -Response $r -ExpectedStatus 200 -UnexpectedSubstrings @('"/test_json":', '"/test_yaml":')

# ==============================================================================
# Summary Report
# ==============================================================================
Write-Banner 'Conduit Test Suite Execution Summary'
Write-Host "  Total Tests: $script:TotalCount" -ForegroundColor Cyan
Write-Host "  Passed:      $script:PassCount" -ForegroundColor Green
if ($script:FailCount -gt 0) {
    Write-Host "  Failed:      $script:FailCount" -ForegroundColor Red
}
else {
    Write-Host '  Failed:      0' -ForegroundColor Green
}
Write-Host "======================================================================`n" -ForegroundColor Cyan

if ($script:FailCount -gt 0) {
    Write-Host "Result: FAILED ($($script:FailCount) test(s) failed)" -ForegroundColor Red
    exit 1
}
else {
    Write-Host 'Result: ALL TESTS PASSED!' -ForegroundColor Green
    exit 0
}
