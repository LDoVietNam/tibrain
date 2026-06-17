@echo off
echo ========================================
echo AUTONOMOUS TASK WORKER - 24/7 AI AGENT
echo ========================================
echo.
echo Configuration:
echo   CLIProxy: http://localhost:1809
echo   Model: Gemini (via CLIProxy tokens)
echo   Interval: 60 seconds
echo.
echo This agent will:
echo   1. Scan task cache every 60s
echo   2. Pick up pending tasks
echo   3. Execute using AI (Gemini/Claude/GPT)
echo   4. Update task status
echo   5. Run 24/7 until stopped
echo.
echo ========================================
echo STARTING WORKER...
echo ========================================
echo.

python autonomous_task_worker.py

echo.
echo ========================================
echo WORKER STOPPED
echo ========================================
pause
