$ErrorActionPreference = 'Stop'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$version = $env:chocolateyPackageVersion

$url64 = "https://github.com/aryanpnd/git-wtm/releases/download/v${version}/git-wtm_windows_amd64.zip"

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  checksum64    = '%CHECKSUM64%'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
