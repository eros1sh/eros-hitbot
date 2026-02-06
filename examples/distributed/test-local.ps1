# Test script for ErosHit Distributed Mode (PowerShell)
# Bu script tek makinede test etmek için kullanılır

param(
    [int]$MasterPort = 8080,
    [int]$WorkerCount = 2,
    [int]$TaskCount = 10
)

$ErrorActionPreference = "Stop"

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     ErosHit Distributed Mode - Local Test                  ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$masterURL = "http://localhost:$MasterPort"
$secretKey = "test-secret-key"

# Function to cleanup processes
function Cleanup {
    Write-Host "`nCleaning up..." -ForegroundColor Yellow
    Get-Process -Name "master" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    Get-Process -Name "worker" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
}

# Cleanup on exit
trap {
    Cleanup
    exit 1
}

# Build binaries
Write-Host "Building binaries..." -ForegroundColor Green
$rootDir = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
cd $rootDir

go build -o bin/master.exe cmd/eroshit/master.go
if ($LASTEXITCODE -ne 0) { throw "Failed to build master" }

go build -o bin/worker.exe cmd/eroshit/worker.go
if ($LASTEXITCODE -ne 0) { throw "Failed to build worker" }

Write-Host "Binaries built successfully" -ForegroundColor Green
Write-Host ""

# Start Master
Write-Host "Starting Master on port $MasterPort..." -ForegroundColor Green
$masterJob = Start-Job -ScriptBlock {
    param($port, $secret)
    cd $using:rootDir
    .\bin\master.exe -bind "127.0.0.1:$port" -secret $secret
} -ArgumentList $MasterPort, $secretKey

Start-Sleep -Seconds 2

# Check if master is running
$masterRunning = $false
$retries = 5
while ($retries -gt 0 -and !$masterRunning) {
    try {
        $response = Invoke-RestMethod -Uri "$masterURL/api/v1/master/status" -Method GET -TimeoutSec 2
        $masterRunning = $true
    } catch {
        Start-Sleep -Seconds 1
        $retries--
    }
}

if (!$masterRunning) {
    Write-Host "Master failed to start!" -ForegroundColor Red
    Receive-Job $masterJob
    exit 1
}

Write-Host "Master is running" -ForegroundColor Green
Write-Host ""

# Start Workers
$workerJobs = @()
for ($i = 1; $i -le $WorkerCount; $i++) {
    Write-Host "Starting Worker $i..." -ForegroundColor Green
    $job = Start-Job -ScriptBlock {
        param($url, $secret, $id)
        cd $using:rootDir
        .\bin\worker.exe -master $url -secret $secret -concurrency 5
    } -ArgumentList $masterURL, $secretKey, $i
    $workerJobs += $job
}

Start-Sleep -Seconds 2

# Submit test tasks
Write-Host ""
Write-Host "Submitting $TaskCount test tasks..." -ForegroundColor Green

for ($i = 1; $i -le $TaskCount; $i++) {
    $body = @{
        url = "https://httpbin.org/get"
        session_id = "test-session-$i"
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$masterURL/api/v1/master/task/submit" `
            -Method POST `
            -ContentType "application/json" `
            -Headers @{ "Authorization" = "Bearer $secretKey" } `
            -Body $body
        Write-Host "  Task $i submitted: $($response.task_id)" -ForegroundColor Gray
    } catch {
        Write-Host "  Failed to submit task $i" -ForegroundColor Red
    }
}

# Monitor progress
Write-Host ""
Write-Host "Monitoring progress..." -ForegroundColor Cyan
Write-Host ""

$stats = @{}
$completed = $false
$monitorTime = 0
$maxMonitorTime = 60  # seconds

while (!$completed -and $monitorTime -lt $maxMonitorTime) {
    Start-Sleep -Seconds 2
    $monitorTime += 2

    try {
        $stats = Invoke-RestMethod -Uri "$masterURL/api/v1/master/stats" `
            -Method GET `
            -Headers @{ "Authorization" = "Bearer $secretKey" }

        $pending = $stats.pending_tasks
        $completed = $stats.completed_tasks
        $failed = $stats.failed_tasks
        $total = $stats.total_tasks
        $workers = $stats.active_workers

        Write-Host "`rWorkers: $workers | Total: $total | Completed: $completed | Failed: $failed | Pending: $pending" -NoNewline

        if ($pending -eq 0 -and ($completed + $failed) -eq $total -and $total -gt 0) {
            $completed = $true
        }
    } catch {
        Write-Host "`rError checking status: $_" -NoNewline -ForegroundColor Red
    }
}

Write-Host ""
Write-Host ""

# Get workers list
try {
    $workers = Invoke-RestMethod -Uri "$masterURL/api/v1/master/workers" `
        -Method GET `
        -Headers @{ "Authorization" = "Bearer $secretKey" }

    Write-Host "Connected Workers:" -ForegroundColor Cyan
    $workers | ForEach-Object {
        Write-Host "  - $($_.id) ($($_.hostname)): $($_.total_tasks) tasks, $($_.success_count) success" -ForegroundColor Gray
    }
} catch {
    Write-Host "Failed to get workers list" -ForegroundColor Red
}

Write-Host ""
Write-Host "Test completed!" -ForegroundColor Green

# Cleanup
Cleanup

# Stop jobs
Stop-Job $masterJob -ErrorAction SilentlyContinue
$workerJobs | ForEach-Object { Stop-Job $_ -ErrorAction SilentlyContinue }

Remove-Job $masterJob -ErrorAction SilentlyContinue
$workerJobs | ForEach-Object { Remove-Job $_ -ErrorAction SilentlyContinue }

Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
