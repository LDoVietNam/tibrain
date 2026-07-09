# === TẠO TUNNEL CLOUDFLARE TRỎ TIBORN MCP SERVER CỘNG 1810 ===
# YÊU CẦU: CHẠY POWERSHELL ỨNG QUYỀN ADMINISTRATOR
# KẾT QUẢ: SẼ RA URL CÔNG KHAI TRỎ ĐẾN http://localhost:1810 (HOẶC BÁO LỖI CHI TIẾT)

try {
    Write-Host "`n🚀 BẮT ĐẦU TẠO TUNNEL CHO CỘNG 1810..." -ForegroundColor Cyan

    # 1. KIỂM TRA VÀ ĐẢM BẢO TIBORN CHẠY TRÊN CỘNG 1810
    Write-Host "🔍 KIỂM TRA TIBORN TRÊN CỘNG 1810..." -ForegroundColor Cyan

    $portInfo = netstat -ano | findstr ":1810"

    if ($portInfo -match "LISTENING") {
        Write-Host "✅ TIBORN ĐANG CHẠY SẴN SÀNG TRÊN CỘNG 1810" -ForegroundColor Green
    } else {
        Write-Host "⚠️ TIBORN CHƯA CHẠY TRÊN CỘNG 1810 - ĐANG KHỞI ĐỘNG BẰNG 'make' (THEO CLAUDE.MD)..." -ForegroundColor Yellow

        pushd "Z:\01_PROJECTS\apps\tibrain" | Out-Null

        $makeProcess = Start-Process -FilePath "cmd" -ArgumentList "/c make" -RedirectStandardOutput "make.log" -RedirectStandardError "make.err" -NoNewWindow -PassThru

        $timeout = [datetime]::Now.AddSeconds(45)
        $tibornReady = $false

        while ([datetime]::Now -lt $timeout) {
            Start-Sleep -Seconds 3

            $portCheck = netstat -ano | findstr ":1810"
            if ($portCheck -match "LISTENING") {
                $tibornReady = $true
                Write-Host "✅ TIBORN ĐÃ KHỞI ĐỘNG THÀNH CÔNG TRÊN CỘNG 1810" -ForegroundColor Green
                break
            }

            if (Test-Path "make.err") {
                $makeError = Get-Content "make.err" -Tail 5
                if ($makeError -match "error|fail|panic|exit status 1"i) {
                    throw "TIBORN KHỎI ĐỘNG THẤT BẢI. LỖI CHI TIẾT: $makeError"
                }
            }
        }

        popd | Out-Null

        if (-not $tibornReady) {
            $finalError = Get-Content "make.err" -Raw -ErrorAction SilentlyContinue
            throw "TIBORN KHỎI ĐỘNG THẤT BẢI SAU 45 GIÂY. KIỂM TRA FILE LỖI: Z:\01_PROJECTS\apps\tibrain\make.err. LỖI GỢI Ý: $($finalError.Trim())"
        }
    }

    # 2. DỪNG TIỀN TRÌNH CLOUDFLARED CŨ (NẸU CÓ)
    Write-Host "`n☁️ ĐANG CHUẨN BỊ TẠO CLOUDFLARE TUNNEL..." -ForegroundColor Cyan
    Get-Process cloudflared -ErrorAction SilentlyContinue | Stop-Process -Force

    # 3. TẠO TUNNEL TRỎ ĐẾN http://localhost:1810
    $tunnelProcess = Start-Process -FilePath "cloudflared" -ArgumentList "tunnel", "--url", "http://localhost:1810" -RedirectStandardOutput "tunnel.log" -RedirectStandardError "tunnel.err" -NoNewWindow -PassThru

    # ĐỢI URL TUNNEL (TỐI ĐA 25 GIÂY)
    Write-Host "⏳ ĐANG CHỢ URL TUNNEL..." -ForegroundColor Cyan
    $timeout = [datetime]::Now.AddSeconds(25)
    $tunnelUrl = $null

    while ([datetime]::Now -lt $timeout) {
        Start-Sleep -Seconds 2
        if (Test-Path "tunnel.log") {
            $lastLines = Get-Content "tunnel.log" -Tail 5
            if ($lastLines -match "https://[a-zA-Z0-9-]+?\.trycloudflare\.com") {
                $tunnelUrl = $matches[0]
                break
            }
        }
    }

    Stop-Process -Id $tunnelProcess.Id -Force -ErrorAction SilentlyContinue

    # 4. TRẢ VỀ KẾT QUẢ CUỐI CÙNG
    if ($tunnelUrl) {
        Write-Host "`n🎉🎉🎉 HOÀN THÀNH! 🎉🎉🎉" -ForegroundColor Green
        Write-Host "   🔗 URL CÔNG KHAI (TRỎ ĐẾN MCP SERVER CỦA TIBRAIN TRÊN CỘNG 1810): $tunnelUrl" -ForegroundColor Yellow
        Write-Host "   🎯 TRỎ ĐẾN: http://localhost:1810" -ForegroundColor Yellow
        Write-Host "`n📝 LƯU Ý QUAN TRỌNG:" -ForegroundColor Cyan
        Write-Host "   - CỬA SỔ POWERSHELL NÀY PHẢI LUÔN MỞ ĐỂ TUNNEL HOẠT ĐỘNG (ĐÓNG CỬA SỔ SẼ DỪNG TUNNEL)" -ForegroundColor White
        Write-Host "   - ĐỂ CHAY VĨNH VIỆN: CHAY LỆNH 'cloudflared service install' SAU NÀY" -ForegroundColor White
        Write-Host "   - URL SẼ THAY ĐỔI MỖI LẦN CHAY TUNNEL MỚI (ĐÂY LÀ TÍNH NĂNG BÌNH THƯỜNG CỦA TRYCLOUDFLARE.COM)" -ForegroundColor White
    } else {
        $tunnelErr = Get-Content "tunnel.err" -Raw -ErrorAction SilentlyContinue
        throw "KHÔNG THỂ LẤY URL TUNNEL SAU 25 GIÂY. LỖI: $($tunnelErr.Trim())"
    }
}
catch {
    Write-Host "`n❌❌❌ LỖI: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "`n🔧 HƯỚNG DẪN KHẮC PHỤC:" -ForegroundColor Yellow
    if ($_.Exception.Message -like "*cloudflared*not*found*") {
        Write-Host "   1. Bạn CHƯA CÀI ĐẶT CLOUDFLARED TUNNEL" -ForegroundColor White
        Write-Host "   2. TẢI TẠI: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/" -ForegroundColor White
        Write-Host "   3. GIẢI ÁP VÀ THÊM THƯ MỤC CHỨA cloudflared.exe VÀO PATH HỆ THỐNG" -ForegroundColor White
    }
    elseif ($_.Exception.Message -like "*TIBORN*KHỎI ĐỘNG*THẤT BẢI*") {
        Write-Host "   1. KIỂM TRA FILE LỖI: Z:\01_PROJECTS\apps\tibrain\make.err" -ForegroundColor White
        Write-Host "   2. CÓ THỂ THIẾU DEPENDENCIES (GO, MAKE, V.V.) - THỬ CHAY 'make' TRỰC TIẾP TRONG THƯ MỤC TIBRAIN TRƯỚC" -ForegroundColor White
        Write-Host "   3. THEO CLAUDE.MD: LUÔN DUNG 'make' - KHÔNG DUNG 'go build' TRỰC TIẾP" -ForegroundColor White
    }
    else {
        Write-Host "   1. ĐAM BAO BAN CHAY POWERSHELL ỨNG QUYỀN ADMINISTRATOR (CLICK CHUỘT PHẢI → 'RUN AS ADMINISTRATOR')" -ForegroundColor White
        Write-Host "   2. KIEM TRA KET NOI MANG VA QUYEN TRUY CAP DEN CLOUDFLARE (KIEM TRA https://api.cloudflare.com/client/v4/accounts)" -ForegroundColor White
        Write-Host "   3. NEU VAI VAN LOI, SAO CHÉP TOÀN BÌN LỖI TRÊN VÀ GỬI LẠI CHO TÔI - TÔI SỖ SỬA LỖI CỤ THỂ" -ForegroundColor White
    }
}
finally {
    # DỌN DẸP FILE TẠM (TÙY CHỌN)
    Remove-Item -Path "make.log", "make.err", "tunnel.log", "tunnel.err" -ErrorAction SilentlyContinue
}

Write-Host "`n💡 NHẤN ENTER ĐỂ THOÁT (HOẶC ĐỂ CỬA SỔ MỞ ĐỂ GIỮ TUNNEL CHẠY)" -ForegroundColor DarkGray
[void][System.Console]::ReadKey($true)