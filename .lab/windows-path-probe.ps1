# Diagnostic only. No environment dumps, credential reads or workflow execution.
$ErrorActionPreference = 'Stop'
Write-Output '=== Windows agent path diagnostic, not Colorama workflow proof ==='
& buildkite-agent artifact download 'windows-path-probe.exe' . --step 'path-probe-build'
if ($LASTEXITCODE -ne 0) { throw 'Cannot download diagnostic executable' }
& (Join-Path (Get-Location) 'windows-path-probe.exe')
if ($LASTEXITCODE -ne 0) { throw 'Diagnostic executable failed' }

$agentPath = (Get-Command buildkite-agent -CommandType Application | Select-Object -First 1).Source
$fsutil = Join-Path ([Environment]::SystemDirectory) 'fsutil.exe'
Write-Output ('PowerShell application path: ' + $agentPath)
for ($path = $agentPath; $path; $path = Split-Path -Parent $path) {
    Write-Output ('=== Executable path component: ' + $path + ' ===')
    try {
        Get-Item -LiteralPath $path -Force |
            Select-Object FullName, Attributes, LinkType, Target |
            ConvertTo-Json -Compress -Depth 3 | Write-Output
    } catch {
        Write-Output ('Get-Item error: ' + $_.Exception.Message)
    }
    $ErrorActionPreference = 'Continue'
    & $fsutil reparsepoint query $path 2>&1 | ForEach-Object { Write-Output $_.ToString() }
    $ErrorActionPreference = 'Stop'
    if ($path -eq [IO.Path]::GetPathRoot($path)) { break }
}

$programFiles = [Environment]::GetFolderPath([Environment+SpecialFolder]::ProgramFiles)
$tools = @(
    (Join-Path $programFiles 'Git\usr\bin\tar.exe'),
    (Join-Path $programFiles 'zstd\zstd.exe'),
    (Join-Path $programFiles 'Git\usr\bin\zstd.exe'),
    'C:\tools\zstd\zstd.exe',
    (Join-Path ([Environment]::SystemDirectory) 'tar.exe')
)
foreach ($tool in $tools) {
    Write-Output ('=== Trusted absolute tool: ' + $tool + ' ===')
    if (Test-Path -LiteralPath $tool -PathType Leaf) {
        Get-Item -LiteralPath $tool |
            Select-Object FullName, Length, @{Name='FileVersion';Expression={$_.VersionInfo.FileVersion}} |
            ConvertTo-Json -Compress | Write-Output
        & $tool --version 2>&1 | ForEach-Object { Write-Output $_.ToString() }
    } else {
        Write-Output 'Absent'
    }
}
Write-Output '=== Diagnostic collection complete ==='
exit 0
