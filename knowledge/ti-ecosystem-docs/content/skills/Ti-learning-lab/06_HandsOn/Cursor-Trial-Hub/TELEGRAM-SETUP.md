# 📞 TELEGRAM SETUP GUIDE

**Purpose:** Get notified when trial resets, services status, errors

---

## ⚡ **QUICK SETUP (2 PHÚT):**

### **BƯỚC 1: TẠO BOT**

```
1. Open Telegram
2. Search: @BotFather
3. Send: /newbot
4. Choose name: AI Hub Bot
5. Choose username: AIHubNotifier_bot
6. Copy token
```

**Example:**
```
Done! Congratulations on your new bot.
You will find it at t.me/AIHubNotifier_bot

Use this token to access the HTTP API:
1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

Keep your token secure!
```

---

### **BƯỚC 2: GET CHAT ID**

```
1. Message your new bot (start conversation)
2. Visit: https://api.telegram.org/botYOUR_TOKEN/getUpdates
3. Look for "chat" -> "id"
```

**Example Response:**
```json
{
  "ok": true,
  "result": [{
    "message": {
      "chat": {
        "id": 123456789,
        "first_name": "Your Name",
        "type": "private"
      }
    }
  }]
}
```

**Copy the ID:** `123456789`

---

### **BƯỚC 3: UPDATE CONFIG**

```powershell
# Edit this file:
notepad C:\Cursor-Trial-Hub\Tokens\telegram.env

# Add your tokens:
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
TELEGRAM_CHAT_ID=123456789

# Save
```

---

## ✅ **TEST TELEGRAM:**

```powershell
# Test notification
cd C:\Cursor-Trial-Hub\Tools
.\send-telegram.ps1 -Message "AI Hub online!"
```

**You should receive:**
```
AI Hub online!
```

---

## 📊 **NOTIFICATIONS:**

```
✅ Trial Reset: When auto-reset runs
✅ Service Status: When services start/stop
✅ Errors: When something goes wrong
```

---

## 🔧 **TROUBLESHOOTING:**

### **Not receiving messages:**

```
1. Check bot token is correct
2. Check chat ID is correct
3. Make sure you messaged the bot first
4. Check TELEGRAM_ENABLED=true
```

### **Bot not responding:**

```
1. Check token format: 1234567890:ABCdef...
2. No spaces in token
3. Bot must be started (send /start)
```

---

## 🎯 **NEXT:**

1. ✅ Create bot (@BotFather)
2. ✅ Get token
3. ✅ Get chat ID
4. ✅ Update telegram.env
5. ✅ Test notification

---

**READY TO GET NOTIFIED!** 📞
