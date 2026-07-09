$tr = 'C:\PROGRA~2\cloudflared\cloudflared.exe tunnel run --token eyJhIjoiNjk3ZmM0YzM1MjdiYTU0MGI0YTU5ODYxMmVlYzkzZDQiLCJ0IjoiYTFiMjA3YTItOTgzZC00YjQ4LWIyNmUtNzJkNjY3MjQ3MmRmIiwicyI6IlpqZGxPRFZoWlRNdE5URXlOaTAwWkdSaExUZzNNR1F0TW1Sa05UaGhOalpqWWpSayJ9'
& C:\Windows\System32\schtasks.exe /create /tn "TiBrainTunnel" /tr $tr /sc ONCE /sd 01/01/2026 /st 00:00 /f
