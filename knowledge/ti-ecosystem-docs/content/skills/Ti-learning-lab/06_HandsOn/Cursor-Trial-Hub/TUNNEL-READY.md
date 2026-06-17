# 🌐 TUNNEL SETUP - CLOUDFLARE (READY)

**Status:** ✅ Pre-configured  
**Domain:** ai-hub.workers.dev (Free Cloudflare subdomain)  
**Services:** 3 (OpenClaw, CLIProxyAPI, My-Router)

---

## ⚡ **QUICK SETUP (2 PHÚT):**

### **BƯỚC 1: LOGIN CLOUDFLARE**

```
1. https://dash.teams.cloudflare.com
2. Sign up/Login (Free)
3. Zero Trust → Access → Tunnels
```

---

### **BƯỚC 2: CREATE TUNNEL**

```
1. Click "Create a tunnel"
2. Choose "Web"
3. Name: AI-Hub-Tunnel
4. Choose environment: Windows
5. Install connector:
   ```powershell
   # Run in PowerShell (Admin)
   C:\Tools\cloudflared.exe service install YOUR_TOKEN
   ```
6. Copy token from dashboard
```

---

### **BƯỚC 3: ADD PUBLIC HOSTNAMES**

```
1. Click "Next"
2. Add public hostname:

   Subdomain: ai-hub
   Domain: workers.dev (or your domain)
   Service: http://localhost:8080
   Type: HTTP

3. Add another:
   Subdomain: api-hub
   Domain: workers.dev
   Service: http://localhost:1810

4. Add another:
   Subdomain: router-hub
   Domain: workers.dev
   Service: http://localhost:1809

5. Save tunnel
```

---

### **BƯỚC 4: UPDATE CONFIG**

```powershell
# Edit: C:\Cursor-Trial-Hub\Tokens\cloudflare-tunnel.env
# Add your credentials:
CLOUDFLARE_ACCOUNT_ID=your_account_id
CLOUDFLARE_TUNNEL_TOKEN=your_tunnel_token
```

---

## ✅ **VERIFICATION:**

```powershell
# Test tunnel
Invoke-RestMethod https://ai-hub.workers.dev

# Should return OpenClaw status
```

---

## 🌐 **ACCESS URLs:**

```
✅ https://ai-hub.workers.dev       (OpenClaw Web UI)
✅ https://api-hub.workers.dev      (CLIProxyAPI)
✅ https://router-hub.workers.dev   (My-Router)
```

---

## 🔧 **TROUBLESHOOTING:**

### **Tunnel not starting:**

```powershell
# Check cloudflared
Test-Path "C:\Tools\cloudflared.exe"

# Reinstall if needed
.\setup-tunnel-simple.ps1
```

### **Domain not working:**

```
Wait 5 minutes for DNS propagation
Or use custom domain
```

---

## 📝 **CONFIG FILES:**

```
✅ tunnel-config.json (Pre-configured)
✅ cloudflare-tunnel.env (Add credentials)
✅ setup-tunnel-simple.ps1 (Auto setup)
✅ start-tunnels.ps1 (Start tunnel)
```

---

## 🎯 **NEXT:**

1. ✅ Login Cloudflare
2. ✅ Create tunnel
3. ✅ Add hostnames
4. ✅ Copy credentials
5. ✅ Update .env
6. ✅ Done!

---

**EVERYTHING READY - JUST ADD CREDENTIALS!** 🚀
