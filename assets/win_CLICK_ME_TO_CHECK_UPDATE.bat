@echo OFF
:: This batch file exists to run win_updater.ps1 without hassle
pushd %~dp0
set updater_script="%~dp0\win_updater.ps1"
powershell -noprofile -nologo -executionpolicy bypass -File %updater_script%