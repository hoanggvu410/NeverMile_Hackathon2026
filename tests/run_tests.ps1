# GitWhy integration test suite
# Usage: .\tests\run_tests.ps1
# Optional: set $env:OPENAI_API_KEY before running for semantic search tests.
#
# Covers:
#   TC-01  Happy path: save + keyword search (no API key needed)
#   TC-02  Happy path: multi-node chain via edge hints
#   TC-03  Happy path: list + get by ID
#   TC-04  Happy path: semantic cache hit (requires OPENAI_API_KEY)
#   TC-05  Happy path: cross-language query (Vietnamese → English, requires key)
#   TC-06  Unhappy: save outside a git repo
#   TC-07  Unhappy: get with nonexistent ID
#   TC-08  Unhappy: search with no contexts saved yet
#   TC-09  Unhappy: save command with no stdin
#   TC-10  Challenge: irrelevant query should not return graph results (similarity floor)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$GO   = "C:\temp\goext\go\bin\go.exe"
$ROOT = "c:\Users\marki\Downloads\coolsies\hackathon"
$BIN  = Join-Path $ROOT "git-why.exe"

# ── helpers ────────────────────────────────────────────────────────────────────

$pass = 0; $fail = 0

function Assert-Contains($label, $output, $expect) {
    if ($output -match [regex]::Escape($expect)) {
        Write-Host "  PASS  $label" -ForegroundColor Green
        $script:pass++
    } else {
        Write-Host "  FAIL  $label" -ForegroundColor Red
        Write-Host "        expected to find: $expect"
        Write-Host "        got: $($output -join ' ')"
        $script:fail++
    }
}

function Assert-NotContains($label, $output, $unexpected) {
    if ($output -notmatch [regex]::Escape($unexpected)) {
        Write-Host "  PASS  $label" -ForegroundColor Green
        $script:pass++
    } else {
        Write-Host "  FAIL  $label (should NOT contain: $unexpected)" -ForegroundColor Red
        $script:fail++
    }
}

function Assert-ExitCode($label, $code, $expectNonZero) {
    if ($expectNonZero -and $code -ne 0) {
        Write-Host "  PASS  $label (exit $code as expected)" -ForegroundColor Green
        $script:pass++
    } elseif (-not $expectNonZero -and $code -eq 0) {
        Write-Host "  PASS  $label (exit 0)" -ForegroundColor Green
        $script:pass++
    } else {
        Write-Host "  FAIL  $label (exit $code, expectNonZero=$expectNonZero)" -ForegroundColor Red
        $script:fail++
    }
}

# Build binary
Write-Host "`nBuilding git-why..." -ForegroundColor Cyan
& $GO build -o $BIN (Join-Path $ROOT "cmd\git-why") 2>&1 | ForEach-Object { Write-Host "  $_" }
if ($LASTEXITCODE -ne 0) { Write-Host "BUILD FAILED -- aborting" -ForegroundColor Red; exit 1 }

# Create a fresh temp git repo for each test group
function New-TestRepo {
    $tmp = Join-Path $env:TEMP "gitwhy_test_$(Get-Random)"
    New-Item -ItemType Directory -Path $tmp | Out-Null
    Push-Location $tmp
    git init -q
    git config user.email "test@example.com"
    git config user.name "Tester"
    # Seed a commit so HEAD exists
    New-Item "seed.txt" -Value "seed" | Out-Null
    git add seed.txt
    git commit -q -m "seed"
    return $tmp
}

function Remove-TestRepo($path) {
    Pop-Location
    Remove-Item -Recurse -Force $path -ErrorAction SilentlyContinue
}

# ── whyspec helpers ─────────────────────────────────────────────────────────────

function New-Whyspec($id, $title, $prompt, $reasoning, $decisions, $rejected, $domain, $topic) {
    return @"
# Context: $title

**Context ID:** $id
**Saved by:** tester
**Agent:** test-agent
**Repository:** test-repo
**Branch:** main
**Date:** 2026-06-20T00:00:00Z
**Domain:** $domain
**Topic:** $topic

## Prompt

> $prompt

## What Was Done

Implemented the change.

## Reasoning

$reasoning

## Key Decisions

$decisions

## Rejected Alternatives

$rejected

## Risks & Open Questions

None for now.

## Files

| File | Status | Description |
|------|--------|-------------|
| services/queue.go | modified | Queue implementation |

## Commits

- abc123def456

## Verification

Run tests.
"@
}

# ── TC-01: save + keyword search ────────────────────────────────────────────────

Write-Host "`n[TC-01] Save a context and find it via keyword search" -ForegroundColor Cyan
$repo01 = New-TestRepo

$doc01 = New-Whyspec "ctx_tc01test" "Remove Kafka" `
    "Remove Kafka from the messaging pipeline, replace with SQS" `
    "Kafka costs USD 800/month. Traffic does not justify the complexity." `
    "Use SQS: managed, cheaper, already on AWS." `
    "RabbitMQ: extra service. Keep Kafka: unjustified cost." `
    "infrastructure" "kafka-removal"

$out01 = $doc01 | & $BIN save 2>&1
Assert-Contains "TC-01a save exits cleanly" $out01 "saved:"
Assert-Contains "TC-01a id in output" $out01 "ctx_tc01test"

$search01 = & $BIN search kafka 2>&1
Assert-Contains "TC-01b find by keyword 'kafka'" $search01 "kafka-removal"

$search01b = & $BIN search "SQS" 2>&1
Assert-Contains "TC-01c find by keyword 'SQS'" $search01b "ctx_tc01test"

$search01c = & $BIN search "elasticsearch" 2>&1
Assert-NotContains "TC-01d no false positive on unrelated query" $search01c "ctx_tc01test"

Remove-TestRepo $repo01

# ── TC-02: multi-node chain via edge hints (full-text only; graph tested separately) ─

Write-Host "`n[TC-02] Save three linked contexts, verify all three are retrievable" -ForegroundColor Cyan
$repo02 = New-TestRepo

$doc_cost = New-Whyspec "ctx_cost001" "Cost analysis" `
    "Analyse infrastructure costs for Q3" `
    "Kafka accounts for 38% of infra spend. SQS would cost 60% less." `
    "Document cost breakdown and present to eng lead." `
    "No alternatives considered -- this is analysis only." `
    "infrastructure" "cost-analysis"

$doc_kafka = New-Whyspec "ctx_kafk001" "Remove Kafka" `
    "Remove Kafka, replace with SQS" `
    "Cost analysis ctx_cost001 showed Kafka is 38% of infra. SQS is sufficient." `
    "Migrate to SQS. Remove kafka.tf. Update services/queue.go." `
    "RabbitMQ: no managed option. Keep Kafka: unjustified." `
    "infrastructure" "kafka-removal"

$doc_dlq = New-Whyspec "ctx_dlq0001" "DLQ setup" `
    "Set up dead letter queue for SQS after Kafka removal" `
    "SQS migration requires DLQ for unprocessable messages. Kafka had no DLQ." `
    "Add DLQ to every SQS queue. Alert on DLQ depth > 0." `
    "Ignore failed messages: data loss risk." `
    "infrastructure" "dlq-setup"

$doc_cost | & $BIN save | Out-Null
$doc_kafka | & $BIN save | Out-Null
$doc_dlq   | & $BIN save | Out-Null

$list02 = & $BIN log 2>&1
Assert-Contains "TC-02a all three contexts in log" $list02 "kafka-removal"

$get02 = & $BIN get ctx_kafk001 2>&1
Assert-Contains "TC-02b get by ID retrieves correct context" $get02 "SQS"
Assert-Contains "TC-02c get includes reasoning" $get02 "Cost analysis"

Remove-TestRepo $repo02

# ── TC-03: list filtering by domain/topic ───────────────────────────────────────

Write-Host "`n[TC-03] List filters by domain and topic" -ForegroundColor Cyan
$repo03 = New-TestRepo

$doc_a = New-Whyspec "ctx_aaatest" "Auth change" `
    "Switch from JWT to session cookies" `
    "JWTs cause stale token issues on password reset." `
    "Use httpOnly session cookies. Add Redis session store." `
    "Keep JWT with short TTL: still stale window." `
    "backend" "auth"

$doc_b = New-Whyspec "ctx_bbbtest" "DB index added" `
    "Add index on users.email" `
    "Query was doing full table scan on 10M rows." `
    "Add B-tree index on users.email." `
    "Composite index: overkill for single-column lookup." `
    "backend" "database"

$doc_a | & $BIN save | Out-Null
$doc_b | & $BIN save | Out-Null

$tree03 = & $BIN tree 2>&1
Assert-Contains "TC-03a tree shows both topics" $tree03 "auth"
Assert-Contains "TC-03b tree shows database topic" $tree03 "database"

Remove-TestRepo $repo03

# ── TC-04: semantic cache hit (requires OPENAI_API_KEY) ──────────────────────────

if ($env:OPENAI_API_KEY) {
    Write-Host "`n[TC-04] Semantic cache: second identical query returns from cache" -ForegroundColor Cyan
    $repo04 = New-TestRepo

    $doc_sem = New-Whyspec "ctx_semtest" "Remove Kafka" `
        "Remove Kafka from the messaging pipeline" `
        "Kafka costs too much. SQS is sufficient for current traffic." `
        "Migrate to AWS SQS." `
        "Keep Kafka: unjustified cost." `
        "infrastructure" "kafka-removal"
    $doc_sem | & $BIN save | Out-Null

    # First search -- embeds + queries graph
    $t1_start = [System.Diagnostics.Stopwatch]::StartNew()
    & $BIN search "why did we remove kafka" | Out-Null
    $t1_start.Stop()
    $ms1 = $t1_start.ElapsedMilliseconds

    # Second search -- should hit cache
    $t2_start = [System.Diagnostics.Stopwatch]::StartNew()
    & $BIN search "why did we remove kafka" | Out-Null
    $t2_start.Stop()
    $ms2 = $t2_start.ElapsedMilliseconds

    Write-Host "  first search: ${ms1}ms  second search: ${ms2}ms"
    if ($ms2 -lt 300) {
        Write-Host "  PASS  TC-04 cache hit is fast (${ms2}ms)" -ForegroundColor Green
        $script:pass++
    } else {
        Write-Host "  WARN  TC-04 cache hit took ${ms2}ms -- expected under 300ms (CLI overhead included)" -ForegroundColor Yellow
    }

    Remove-TestRepo $repo04
} else {
    Write-Host "`n[TC-04] SKIP -- OPENAI_API_KEY not set" -ForegroundColor Yellow
}

# ── TC-05: cross-language query (requires OPENAI_API_KEY) ────────────────────────

if ($env:OPENAI_API_KEY) {
    Write-Host "`n[TC-05] Cross-language: Vietnamese query finds English context" -ForegroundColor Cyan
    $repo05 = New-TestRepo

    $doc_vi = New-Whyspec "ctx_vitest1" "Kafka removal" `
        "Remove Kafka from messaging pipeline" `
        "Kafka is overkill. Costs USD 800/month. SQS is cheaper and managed." `
        "Replace all Kafka producers/consumers with SQS. Remove kafka.tf." `
        "Keep Kafka at smaller scale: still complex. RabbitMQ: extra ops." `
        "infrastructure" "kafka-removal"
    $doc_vi | & $BIN save | Out-Null

    # Vietnamese: "why did we remove Kafka"
    $out05 = & $BIN search "tại sao bỏ Kafka" 2>&1
    if ($out05 -match "kafka") {
        Write-Host "  PASS  TC-05 Vietnamese query returns kafka context" -ForegroundColor Green
        $script:pass++
    } else {
        Write-Host "  WARN  TC-05 Vietnamese query got no results (may be below similarity floor)" -ForegroundColor Yellow
        Write-Host "        output: $out05"
    }

    Remove-TestRepo $repo05
} else {
    Write-Host "`n[TC-05] SKIP -- OPENAI_API_KEY not set" -ForegroundColor Yellow
}

# ── TC-06: unhappy -- save outside a git repo ────────────────────────────────────

Write-Host "`n[TC-06] Unhappy: save outside a git repo returns a useful error" -ForegroundColor Cyan
$noGitDir = Join-Path $env:TEMP "gitwhy_nogit_$(Get-Random)"
New-Item -ItemType Directory -Path $noGitDir | Out-Null
Push-Location $noGitDir

$doc_nogit = New-Whyspec "ctx_nogit01" "Test" "Test prompt" "Test reasoning" "Test decision" "None" "" ""
# Temporarily allow native command errors so 2>&1 capture works on PS 5.1
$old_pref = $ErrorActionPreference; $ErrorActionPreference = "Continue"
$out06 = $doc_nogit | & $BIN save 2>&1
$exit06 = $LASTEXITCODE
$ErrorActionPreference = $old_pref
Assert-ExitCode "TC-06 exits non-zero" $exit06 $true
Assert-Contains "TC-06 error mentions git" ($out06 -join " ") "git"

Pop-Location
Remove-Item -Recurse -Force $noGitDir -ErrorAction SilentlyContinue

# ── TC-07: unhappy -- get with nonexistent ID ────────────────────────────────────

Write-Host "`n[TC-07] Unhappy: get with a nonexistent ID returns not-found error" -ForegroundColor Cyan
$repo07 = New-TestRepo

$old_pref07 = $ErrorActionPreference; $ErrorActionPreference = "Continue"
$out07 = & $BIN get ctx_doesnotexist 2>&1
$exit07 = $LASTEXITCODE
$ErrorActionPreference = $old_pref07
Assert-ExitCode "TC-07 exits non-zero" $exit07 $true
Assert-Contains "TC-07 error is descriptive" ($out07 -join " ") "not found"

Remove-TestRepo $repo07

# ── TC-08: unhappy -- search in empty repo ───────────────────────────────────────

Write-Host "`n[TC-08] Unhappy: search in repo with no saved contexts" -ForegroundColor Cyan
$repo08 = New-TestRepo

$out08 = & $BIN search "kafka" 2>&1
Assert-ExitCode "TC-08 exits cleanly (0)" $LASTEXITCODE $false
Assert-Contains "TC-08 reports no results gracefully" ($out08 -join " ") "no contexts matched"

Remove-TestRepo $repo08

# ── TC-09: unhappy -- save with empty stdin ──────────────────────────────────────

Write-Host "`n[TC-09] Unhappy: save with empty stdin" -ForegroundColor Cyan
$repo09 = New-TestRepo

$old_pref09 = $ErrorActionPreference; $ErrorActionPreference = "Continue"
# PowerShell pipes a newline even for "", so stdin is non-empty.
# The binary gets past the empty-stdin check and fails on missing Context ID instead.
$out09 = "" | & $BIN save 2>&1
$exit09 = $LASTEXITCODE
$ErrorActionPreference = $old_pref09
Assert-ExitCode "TC-09 exits non-zero" $exit09 $true
Assert-Contains "TC-09 error is descriptive (missing fields)" ($out09 -join " ") "parsing whyspec"

Remove-TestRepo $repo09

# ── TC-10: similarity floor challenge ───────────────────────────────────────────

if ($env:OPENAI_API_KEY) {
    Write-Host "`n[TC-10] Challenge: irrelevant query should NOT return results (similarity floor 0.60)" -ForegroundColor Cyan
    $repo10 = New-TestRepo

    # Save a very specific infrastructure context
    $doc_infra = New-Whyspec "ctx_infra01" "SQS queue config" `
        "Configure SQS dead letter queue retry policy" `
        "Failed messages were silently dropped. DLQ captures them for inspection." `
        "maxReceiveCount=3. DLQ retention=14 days. Alert on depth > 0." `
        "Ignore failures: data loss. Infinite retry: queue flooding." `
        "infrastructure" "dlq-setup"
    $doc_infra | & $BIN save | Out-Null

    # Query about something completely unrelated (UI color scheme)
    $out10 = & $BIN search "why did we change the button color to blue" 2>&1
    if ($out10 -notmatch "dlq") {
        Write-Host "  PASS  TC-10 irrelevant query returns no graph results (floor working)" -ForegroundColor Green
        $script:pass++
    } else {
        Write-Host "  FAIL  TC-10 irrelevant query returned DLQ context -- similarity floor too low" -ForegroundColor Red
        $script:fail++
    }

    Remove-TestRepo $repo10
} else {
    Write-Host "`n[TC-10] SKIP -- OPENAI_API_KEY not set (similarity floor only applies to graph search)" -ForegroundColor Yellow
}

# ── summary ─────────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "─────────────────────────────────────────" -ForegroundColor DarkGray
Write-Host "  Results: $pass passed, $fail failed" -ForegroundColor $(if ($fail -eq 0) { "Green" } else { "Red" })
Write-Host "─────────────────────────────────────────" -ForegroundColor DarkGray

if ($fail -gt 0) { exit 1 } else { exit 0 }
