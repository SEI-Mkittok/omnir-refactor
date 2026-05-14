@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "SCRIPT_DIR=%~dp0"
set "TARGET_SCRIPT=%SCRIPT_DIR%pre-push-check.sh"
set "GIT_BASH="

if exist "%ProgramFiles%\Git\bin\bash.exe" (
  set "GIT_BASH=%ProgramFiles%\Git\bin\bash.exe"
)

if not defined GIT_BASH if exist "%ProgramFiles(x86)%\Git\bin\bash.exe" (
  set "GIT_BASH=%ProgramFiles(x86)%\Git\bin\bash.exe"
)

if not defined GIT_BASH (
  for %%I in (git.exe) do set "GIT_CMD=%%~$PATH:I"
  if defined GIT_CMD (
    for %%J in ("!GIT_CMD!") do set "GIT_BASH=%%~dpJ..\bin\bash.exe"
  )
)

if not defined GIT_BASH (
  echo ERROR: Could not find Git Bash.
  echo Install Git for Windows or run this script from Git Bash.
  exit /b 1
)

if not exist "%TARGET_SCRIPT%" (
  echo ERROR: Missing script "%TARGET_SCRIPT%".
  exit /b 1
)

"%GIT_BASH%" "%TARGET_SCRIPT%" %*
exit /b %ERRORLEVEL%
