@echo off
chcp 936 >nul
cd /d "%~dp0"

echo [1/5] 清理旧的资源文件与二进制文件...
del *.syso 2>nul
del RuleCraft.exe 2>nul
del RuleCraft-upx.exe 2>nul
del api\webui\User-Manual.md
del api\webui\User-Manual_CN.md

REM 检查是否因为 exe 正在运行而导致删除失败
if exist RuleCraft.exe goto ERR_DEL
if exist RuleCraft-upx.exe goto ERR_DEL
if exist api\webui\User-Manual.md goto ERR_DEL
if exist api\webui\User-Manual_CN.md goto ERR_DEL

echo [2/5] 生成 Windows 资源文件(.syso)...
go-winres make
copy docs\User-Manual.md api\webui\User-Manual.md
copy docs\使用手册.md api\webui\User-Manual_CN.md

if errorlevel 1 goto ERR_WINRES

echo [3/5] 编译 Go 项目...
go build -ldflags="-H windowsgui" -o RuleCraft.exe .
if errorlevel 1 goto ERR_BUILD

echo [4/5] 清理临时资源文件...
del *.syso 2>nul

echo [5/5] 执行 UPX 压缩...
upx --best -o RuleCraft-upx.exe .\RuleCraft.exe
if errorlevel 1 goto ERR_UPX

echo.
echo ===========================================
echo === BUILD OK: RuleCraft.exe
echo === UPX OK:   RuleCraft-upx.exe
echo ===========================================
pause
exit /b 0

:ERR_DEL
echo ERROR: 无法删除旧的 exe 文件，请先关闭正在运行的 RuleCraft！
pause
exit /b 1

:ERR_WINRES
echo ERROR: go-winres 执行失败，请检查工具是否安装！
pause
exit /b 1

:ERR_BUILD
echo ERROR: Go 编译失败，请检查 Go 代码是否有错！
pause
exit /b 1

:ERR_UPX
echo ERROR: UPX 压缩失败，请检查 UPX 是否加入环境变量！
pause
exit /b 1