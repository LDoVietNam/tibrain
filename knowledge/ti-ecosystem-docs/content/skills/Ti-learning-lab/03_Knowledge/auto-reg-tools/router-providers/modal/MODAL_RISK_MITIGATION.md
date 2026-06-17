# Modal GLM-5 Risk Mitigation Rules

## Key Expiration Risks

### Current Key Status
- **Type**: Trial key (`modalresearch_*`)
- **Free period**: Until **April 30, 2026** (Modal GLM-5 free trial)
- **Estimated individual key expiry**: ~90 days from creation
- **Credits limit**: $30/month (Starter plan) - may exhaust with heavy usage

### Failure Modes

1. **Key Expired** (401 Unauthorized)
   - Solution: Switch to backup key immediately
   - Action: Update `MODAL_API_KEY` in `Z:\00_SECRET\.routerenv`

2. **Credits Exhausted** (403/Rate limit)
   - Solution: Wait for monthly reset or upgrade to paid plan
   - Monitor: Modal dashboard → Billing

3. **Free Tier End** (after April 30, 2026)
   - Solution: Upgrade to paid plan or migrate to self-hosted
   - Timeline: Must be resolved before May 1, 2026

---

## Fallback Strategies

### 1. Dual API Keys (Immediate Protection)
```json
// OpenCode/Kilo config - switch between keys by changing env var
"apiKey": "{env:MODAL_API_KEY}"  // Primary
// OR
"apiKey": "{env:MODAL_API_KEY_2}" // Backup
```

**Action**: Keep both `MODAL_API_KEY` and `MODAL_API_KEY_2` updated with fresh keys.

### 2. Ti Router Backup (Port 1807)
If Modal direct fails, switch OpenCode/Kilo config to use Ti Router:
```json
"baseURL": "http://localhost:1807/v1"
```

Ti Router can be configured to route to different backends.

### 3. Ti Backend Fallback (Port 1806)
Current Ti Backend runs Claude via OpenRouter. Switch Kilo to:
```json
"model": "ti/claude-haiku-4-5-20251001"
```

---

## Monitor & Alert Rules

### Check Credits Weekly
```powershell
# Manual check - Modal CLI
modal billing report --workspace=YOUR_WORKSPACE

# Or check dashboard: https://modal.com/dashboard/billing
```

### Test Keys Daily (Automated)
Create scheduled task to run `test-modal.ps1` daily and alert on 401.

### Watch for These Signals
- **401 Unauthorized** → Key expired/revoked
- **403 Rate limit** → Credits exhausted
- **404 Not Found** → Endpoint changed (unlikely)

---

## Action Plan Timeline

| Date | Action |
|------|--------|
| **Now** | Have 2 fresh keys ready (one primary, one backup) |
| **April 15, 2026** | Renew/refresh keys before free tier ends |
| **April 25, 2026** | Decide: paid Modal plan vs self-host GLM-5 |
| **May 1, 2026** | **Deadline** - free tier ends |

---

## Emergency Procedures

### If Key Fails Mid-Session:
1. Switch `MODAL_API_KEY` → `MODAL_API_KEY_2` in `.routerenv`
2. Reload: `powershell -File "load-env.ps1"`
3. Restart OpenCode/Kilo
4. If both fail → switch Ti Router config → restart

### If All Modal Options Fail:
1. Use Ti Backend (Claude via OpenRouter)
2. Or switch to other provider (OpenRouter, Groq, etc.)
3. Config fallback in `opencode.json`:
```json
{
  "provider": {
    "modal": { ... },
    "openrouter": { ... }  // Backup provider
  }
}
```

---

## Key Rotation Best Practices

1. **Never delete old keys** - keep as fallback
2. **Label keys** in Modal dashboard: `GLM5-Primary-APR2026`
3. **Document key dates**: creation/expiry in this file
4. **Test keys weekly** - automated script

---

## Secrets Management Compliance (RULES.md)

- ✅ All keys stored in `Z:\00_SECRET\.routerenv`
- ✅ Never hardcode in config files
- ✅ Never commit to git
- ✅ Backup before changing
- ❌ **Do NOT delete old keys without confirmation**

---

## Current Keys Log

| Name | Value (partial) | Created | Expires | Status |
|------|----------------|---------|---------|--------|
| MODAL_API_KEY | `modalresearch_z6...` | Unknown | ~90 days | ⚠️ Trial, may expire |
| MODAL_API_KEY_2 | `modalresearch_mj...` | Unknown | ~90 days | ⚠️ Trial, may expire |
| **Action** | **Get 2 new keys** | **ASAP** | **Before April 30** | **🔴 Urgent** |

---

## Next Steps

1. 🔴 **IMMEDIATE**: Get 2 new Modal API keys (non-trial if possible)
2. 🟡 Update `.routerenv` with new keys (keep old as backup)
3. 🟡 Test all CLI tools work with new keys
4. 🟢 Set up weekly key health check script
5. 🟢 Plan migration strategy for post-April 2026

---

**Last Updated**: 2026-04-17  
**Next Review**: 2026-04-20 (weekly)
