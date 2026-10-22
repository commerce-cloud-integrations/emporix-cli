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

function Install-Epx {
  $ErrorActionPreference = "Stop"
  $ProgressPreference = "SilentlyContinue"
  # Windows PowerShell 5.1 may not offer TLS 1.2 by default; GitHub requires it.
  [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

  $repo = "commerce-cloud-integrations/emporix-cli"
  $version = $env:EPX_VERSION
  if (-not $version) {
    $version = (Invoke-RestMethod -UseBasicParsing "https://api.github.com/repos/$repo/releases/latest").tag_name
  }
  if (-not $version.StartsWith("v")) { $version = "v$version" }
  if ($version -notmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$') {
    throw "epx installer: EPX_VERSION is not a release tag such as v0.1.2"
  }

  # The OS architecture, also when this PowerShell runs emulated; older .NET Framework
  # versions lack OSArchitecture, so the environment of Windows answers instead.
  $arch = $null
  try { $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString() } catch { }
  if (-not $arch) { $arch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE } }
  switch ($arch) {
    { $_ -in "X64", "AMD64" } { $cpu = "amd64" }
    { $_ -eq "Arm64" } { $cpu = "arm64" }
    default { throw "epx installer: no epx binary exists for the CPU $arch" }
  }
  $name = "epx_${version}_windows_$cpu"
  $base = "https://github.com/$repo/releases/download/$version"
  $dir = $env:EPX_INSTALL_DIR
  if (-not $dir) { $dir = Join-Path $env:LOCALAPPDATA "Programs\epx\bin" }

  $tmp = Join-Path ([IO.Path]::GetTempPath()) ("epx-install-" + [Guid]::NewGuid())
  New-Item -ItemType Directory -Path $tmp | Out-Null
  try {
    Write-Host "epx installer: downloading epx $version ($name)"
    $zip = Join-Path $tmp "$name.zip"
    $sums = Join-Path $tmp "checksums.txt"
    Invoke-WebRequest -UseBasicParsing "$base/$name.zip" -OutFile $zip
    Invoke-WebRequest -UseBasicParsing "$base/epx_${version}_checksums.txt" -OutFile $sums
    $want = $null
    foreach ($line in Get-Content $sums) {
      $fields = $line -split '\s+'
      if ($fields.Count -ge 2 -and $fields[1] -eq "$name.zip") { $want = $fields[0] }
    }
    if (-not $want) { throw "epx installer: the checksums of $version do not list $name.zip" }
    $sha = [Security.Cryptography.SHA256]::Create()
    $stream = [IO.File]::OpenRead($zip)
    try { $got = -join ($sha.ComputeHash($stream) | ForEach-Object { $_.ToString("x2") }) }
    finally { $stream.Dispose(); $sha.Dispose() }
    if ($got -ne $want) { throw "epx installer: $name.zip does not match its checksum" }

    Expand-Archive -Path $zip -DestinationPath $tmp
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    $exe = Join-Path $dir "epx.exe"
    Copy-Item -Force (Join-Path $tmp "$name\epx.exe") $exe
  }
  finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
  }

  $reported = & $exe version
  if ($LASTEXITCODE -ne 0 -or -not "$reported".StartsWith("epx $version ")) {
    throw "epx installer: $exe reports '$reported', not $version"
  }
  Write-Host "epx installer: installed $reported to $exe"

  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $entries = @($userPath -split ';' | Where-Object { $_ })
  if ($entries -notcontains $dir) {
    if ($env:EPX_NO_MODIFY_PATH -eq "1") {
      Write-Host "epx installer: $dir is not on your Path; add it to use epx from any terminal"
    } else {
      [Environment]::SetEnvironmentVariable("Path", (@($dir) + $entries) -join ';', "User")
      Write-Host "epx installer: added $dir to your Path; new terminals find epx"
    }
  }
  if (@($env:Path -split ';') -notcontains $dir) { $env:Path = "$dir;$env:Path" }
}

Install-Epx
