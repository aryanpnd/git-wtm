$ErrorActionPreference = 'Stop'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

Remove-Item "$toolsDir\git-wtm.exe" -Force -ErrorAction SilentlyContinue
