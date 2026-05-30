@echo OFF
:: 此批处理文件用于方便地运行 win_updater.ps1
pushd %~dp0
set updater_script="%~dp0\win_updater.ps1"
powershell -noprofile -nologo -executionpolicy bypass -File %updater_script%