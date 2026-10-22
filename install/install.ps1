# Installs epx, a command-line client for the Emporix Commerce APIs, on Windows:
#
#   powershell -ExecutionPolicy ByPass -c "irm https://github.com/commerce-cloud-integrations/emporix-cli/releases/latest/download/install.ps1 | iex"
#
# It downloads the release archive of this machine from GitHub, checks it against the
# release's checksums file, puts epx.exe into %LOCALAPPDATA%\Programs\epx\bin and adds that
# directory to the user's Path. No administrator rights are needed.
#
# Settings (environment variables):
#   EPX_VERSION         release tag to install, such as v0.1.2 (default: the latest release)
#   EPX_INSTALL_DIR     directory for epx.exe (default: %LOCALAPPDATA%\Programs\epx\bin)
#   EPX_NO_MODIFY_PATH  set to 1 to leave the user's Path alone
#
# Everything runs inside Install-Epx, so a download cut off midway runs nothing.

$EpxRepo = "commerce-cloud-integrations/emporix-cli"

# Get-EpxVersion returns the release tag to install: EPX_VERSION, or the latest release.
function Get-EpxVersion {
  $version = $env:EPX_VERSION
  if (-not $version) {
    $version = (Invoke-RestMethod -UseBasicParsing "https://api.github.com/repos/$EpxRepo/releases/latest").tag_name
  }
  if (-not $version.StartsWith("v")) { $version = "v$version" }
  if ($version -notmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$') {
    throw "epx installer: EPX_VERSION is not a release tag such as v0.1.2"
  }
  return $version
}

# Get-EpxCpu returns the CPU of the release archive for this machine: amd64 or arm64.
function Get-EpxCpu {
  # The OS architecture, also when this PowerShell runs emulated; older .NET Framework
  # versions lack OSArchitecture, so the environment of Windows answers instead.
  $arch = $null
  try {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
  } catch {
    Write-Verbose "OSArchitecture is not available ($_); reading PROCESSOR_ARCHITECTURE"
  }
  if (-not $arch) { $arch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE } }
  switch ($arch) {
    { $_ -in "X64", "AMD64" } { return "amd64" }
    { $_ -eq "Arm64" } { return "arm64" }
    default { throw "epx installer: no epx binary exists for the CPU $arch" }
  }
}

# Assert-EpxChecksum fails unless the file has the SHA-256 that the checksums file lists
# for its name.
function Assert-EpxChecksum([string]$File, [string]$Sums) {
  $name = Split-Path -Leaf $File
  $want = $null
  foreach ($line in Get-Content $Sums) {
    $fields = $line -split '\s+'
    if ($fields.Count -ge 2 -and $fields[1] -eq $name) { $want = $fields[0] }
  }
  if (-not $want) { throw "epx installer: the checksums file does not list $name" }
  $sha = [Security.Cryptography.SHA256]::Create()
  $stream = [IO.File]::OpenRead($File)
  try { $got = -join ($sha.ComputeHash($stream) | ForEach-Object { $_.ToString("x2") }) }
  finally { $stream.Dispose(); $sha.Dispose() }
  if ($got -ne $want) { throw "epx installer: $name does not match its checksum" }
}

# Add-EpxToPath puts the directory on the user's Path (unless EPX_NO_MODIFY_PATH is 1) and
# on the Path of this session.
function Add-EpxToPath([string]$Dir) {
  $entries = @([Environment]::GetEnvironmentVariable("Path", "User") -split ';' | Where-Object { $_ })
  if ($entries -notcontains $Dir) {
    if ($env:EPX_NO_MODIFY_PATH -eq "1") {
      Write-Output "epx installer: $Dir is not on your Path; add it to use epx from any terminal"
    } else {
      [Environment]::SetEnvironmentVariable("Path", (@($Dir) + $entries) -join ';', "User")
      Write-Output "epx installer: added $Dir to your Path; new terminals find epx"
    }
  }
  if (@($env:Path -split ';') -notcontains $Dir) { $env:Path = "$Dir;$env:Path" }
}

function Install-Epx {
  $ErrorActionPreference = "Stop"
  $ProgressPreference = "SilentlyContinue"
  # Windows PowerShell 5.1 may not offer TLS 1.2 by default; GitHub requires it.
  [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

  $version = Get-EpxVersion
  $name = "epx_${version}_windows_$(Get-EpxCpu)"
  $base = "https://github.com/$EpxRepo/releases/download/$version"
  $dir = $env:EPX_INSTALL_DIR
  if (-not $dir) { $dir = Join-Path $env:LOCALAPPDATA "Programs\epx\bin" }
  $exe = Join-Path $dir "epx.exe"

  $tmp = Join-Path ([IO.Path]::GetTempPath()) ("epx-install-" + [Guid]::NewGuid())
  New-Item -ItemType Directory -Path $tmp | Out-Null
  try {
    Write-Output "epx installer: downloading epx $version ($name)"
    $zip = Join-Path $tmp "$name.zip"
    $sums = Join-Path $tmp "checksums.txt"
    Invoke-WebRequest -UseBasicParsing "$base/$name.zip" -OutFile $zip
    Invoke-WebRequest -UseBasicParsing "$base/epx_${version}_checksums.txt" -OutFile $sums
    Assert-EpxChecksum $zip $sums
    Expand-Archive -Path $zip -DestinationPath $tmp
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    Copy-Item -Force (Join-Path $tmp "$name\epx.exe") $exe
  }
  finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
  }

  $reported = & $exe version
  if ($LASTEXITCODE -ne 0 -or -not "$reported".StartsWith("epx $version ")) {
    throw "epx installer: $exe reports '$reported', not $version"
  }
  Write-Output "epx installer: installed $reported to $exe"
  Add-EpxToPath $dir
}

Install-Epx
