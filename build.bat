@echo off
chcp 936 >nul
cd /d "%~dp0"

set "OUTDIR=build"

echo [1/6] 清理旧的临时/产物文件...
del *.syso 2>nul
del "%OUTDIR%\RuleCraft.exe" 2>nul
del "%OUTDIR%\RuleCraft-upx.exe" 2>nul
del "api\webui\User-Manual.md" 2>nul
del "api\webui\User-Manual_CN.md" 2>nul

if exist "%OUTDIR%\RuleCraft.exe" goto ERR_DEL
if exist "%OUTDIR%\RuleCraft-upx.exe" goto ERR_DEL
if exist "api\webui\User-Manual.md" goto ERR_DEL
if exist "api\webui\User-Manual_CN.md" goto ERR_DEL

echo [2/6] 创建 build 目录结构...
if not exist "%OUTDIR%" mkdir "%OUTDIR%"
if not exist "%OUTDIR%\Input_Plugins" mkdir "%OUTDIR%\Input_Plugins"
if not exist "%OUTDIR%\Output_Plugins" mkdir "%OUTDIR%\Output_Plugins"
if not exist "%OUTDIR%\tasks" mkdir "%OUTDIR%\tasks"
if not exist "%OUTDIR%\logs" mkdir "%OUTDIR%\logs"
if not exist "%OUTDIR%\docs" mkdir "%OUTDIR%\docs"

echo [3/6] 生成 Windows 资源文件(.syso)并准备嵌入手册...
go-winres make
if errorlevel 1 goto ERR_WINRES
copy /Y "docs\User-Manual.md" "api\webui\User-Manual.md" >nul
copy /Y "docs\使用手册.md" "api\webui\User-Manual_CN.md" >nul
if errorlevel 1 goto ERR_WINRES

echo [4/6] 编译 Go 项目 → %OUTDIR%\RuleCraft.exe ...
go build -ldflags="-H windowsgui" -o "%OUTDIR%\RuleCraft.exe" .
if errorlevel 1 goto ERR_BUILD

echo [5/6] 清理临时 .syso ...
del *.syso 2>nul

echo [6/6] 整理运行时文件到 %OUTDIR%\ ...
copy /Y "config.json" "%OUTDIR%\config.json" >nul
xcopy /E /I /Y "Input_Plugins\*" "%OUTDIR%\Input_Plugins\" >nul
if not exist "Output_Plugins" mkdir "Output_Plugins"
xcopy /E /I /Y "Output_Plugins\*" "%OUTDIR%\Output_Plugins\" >nul 2>nul
xcopy /E /I /Y "tasks\*" "%OUTDIR%\tasks\" >nul
copy /Y "docs\User-Manual.md" "%OUTDIR%\docs\User-Manual.md" >nul
copy /Y "docs\使用手册.md" "%OUTDIR%\docs\使用手册.md" >nul
copy /Y "docs\Plugin-Manual.md" "%OUTDIR%\docs\Plugin-Manual.md" >nul
copy /Y "docs\插件开发手册.md" "%OUTDIR%\docs\插件开发手册.md" >nul

echo.
echo [optional] 执行 UPX 压缩...
where upx >nul 2>&1
if errorlevel 1 (
    echo 未检测到 UPX，跳过压缩。普通版已就绪。
) else (
    upx --best -o "%OUTDIR%\RuleCraft-upx.exe" "%OUTDIR%\RuleCraft.exe"
    if errorlevel 1 (
        echo UPX 压缩失败，保留未压缩版。
    ) else (
        echo UPX OK: %OUTDIR%\RuleCraft-upx.exe
    )
)

echo.
echo ===========================================
echo === BUILD OK: %OUTDIR%\RuleCraft.exe
echo === 可带走目录: %OUTDIR%\
echo ===========================================
pause
exit /b 0

:ERR_DEL
echo ERROR: 无法删除旧的 exe 文件，请先关闭正在运行的 RuleCraft。
pause
exit /b 1

:ERR_WINRES
echo ERROR: go-winres / 手册复制失败，请检查工具是否安装。
pause
exit /b 1

:ERR_BUILD
echo ERROR: Go 编译失败，请检查 Go 环境是否有错误。
del *.syso 2>nul
pause
exit /b 1