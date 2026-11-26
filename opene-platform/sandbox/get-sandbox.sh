#!/usr/bin/env bash
set -euo pipefail
# The executable is Windows-only. Git Bash/WSL may invoke PowerShell directly.
here=$(cd "$(dirname "$0")" && pwd)
case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*) script=$(cygpath -w "$here/get-sandbox.ps1");;
  Linux*) command -v wslpath >/dev/null || { echo 'Windows/WSL krävs' >&2; exit 1; }; script=$(wslpath -w "$here/get-sandbox.ps1");;
  *) echo 'Windows/WSL krävs' >&2; exit 1;;
esac
exec powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$script" "$@"
