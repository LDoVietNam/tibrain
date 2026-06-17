---
tags: ["tibrain", "documentation", "skill", "provider-notion", "router"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Notion Database Configuration Guide - CLI Knowledge Sync

> **Purpose**: Configure Notion database to support CLI documentation sync with AI metadata
> **Database ID**: `34fb60e8155a80c787a6ce3ea9f631af`
> **Status**: Configuration Required

---

## 📋 Prerequisites

1. **Notion Access**: Access to Notion database `34fb60e8155a80c787a6ce3ea9f631af`
2. **API Key**: `NOTION_API_KEY` from `Z:\00_SECRET\notion.env`
3. **Permissions**: Database permissions to create pages and modify properties

---

## 🔧 Step 1: Add CLI Architecture Category

### 1.1 Access Database
1. Open Notion
2. Navigate to database `34fb60e8155a80c787a6ce3ea9f631af`
3. Click on the "Category" property (or create if doesn't exist)

### 1.2 Add CLI Architecture Option
1. Click on the Category property
2. Add new option: **"CLI Architecture"**
3. Choose a color (recommended: Blue or Purple)
4. Save changes

### 1.3 Add Subcategories (Optional but Recommended)
If you want to track specific CLI components, add these as separate options:
- **Flag Management** (Green)
- **Error Handling** (Orange)
- **Command Validation** (Yellow)
- **Tool System** (Red)
- **MCP Integration** (Pink)
- **Examples** (Gray)
- **Overview** (Blue)

**Note**: The AI sync agent will use the category mapping to assign these automatically.

---

## 🔧 Step 2: Configure Properties

### 2.1 Required Properties (Already Exist)
These properties should already exist in the database:
- ✅ **Tên** (Title) - Document name
- ✅ **Type** (Select) - Document type (Changelog, Guide, Pattern, Architecture)
- ✅ **Status** (Select) - Document status (Done, In Progress, Pending)
- ✅ **Priority** (Select) - Priority level (High, Medium, Low, P0, P1, P2, P3)

### 2.2 New Properties to Add

#### Property 1: **Folder** (Select)
- **Type**: Select
- **Options**: 
  - CLI (Blue)
  - Router (Green)
  - General (Gray)
  - (Add more as needed)
- **Purpose**: Track which folder the file belongs to

#### Property 2: **File Path** (Text)
- **Type**: Text
- **Purpose**: Store the full file path for reference

#### Property 3: **AI Summary** (Text)
- **Type**: Text
- **Purpose**: Store AI-generated summary of the document

#### Property 4: **Folder Context** (Text)
- **Type**: Text
- **Purpose**: Store folder context (files in the same folder)

#### Property 5: **Related Knowledge** (Multi-select)
- **Type**: Multi-select
- **Options**: (Will be populated dynamically)
- **Purpose**: Link to related knowledge files

#### Property 6: **AI Category** (Select)
- **Type**: Select
- **Options**: Same as Category
- **Purpose**: Store AI-determined category

#### Property 7: **AI Priority** (Select)
- **Type**: Select
- **Options**: P0, P1, P2, P3, High, Medium, Low
- **Purpose**: Store AI-determined priority

#### Property 8: **AI Relationships** (Multi-select)
- **Type**: Multi-select
- **Options**: (Will be populated dynamically)
- **Purpose**: Store AI-detected relationships

#### Property 9: **Sync Status** (Select)
- **Type**: Select
- **Options**: 
  - Synced (Green)
  - Pending (Yellow)
  - Error (Red)
- **Purpose**: Track sync status

#### Property 10: **Last Sync** (Date)
- **Type**: Date
- **Purpose**: Track when the file was last synced

---

## 🔧 Step 3: Verify Property Setup

### 3.1 Check Existing Properties
1. Open database in Notion
2. Click on "Properties" header
3. Verify all required properties exist
4. Verify all new properties are added

### 3.2 Test Property Creation (Optional)
Create a test page to verify all properties work:
1. Click "New" in database
2. Fill in all properties
3. Save and verify no errors

---

## 🔧 Step 4: Test API Access

### 4.1 Verify API Key
```bash
# Check API key exists
cat Z:\00_SECRET\notion.env
```

Expected output:
```
NOTION_API_KEY=secret_xxx...
```

### 4.2 Test Database Access
Run this simple test to verify API access:

```bash
cd /z/10_WORKPLACE/Ti/apps/automation
export NOTION_API_KEY=$(cat Z:\00_SECRET\notion.env)
curl -X GET 'https://api.notion.com/v1/databases/34fb60e8155a80c787a6ce3ea9f631af' \
  -H "Authorization: Bearer $NOTION_API_KEY" \
  -H "Notion-Version: 2022-06-28"
```

Expected response: JSON with database information

---

## 🔧 Step 5: Configure CLI-Specific Folder Option

Since the sync agent will sync from `Ti-learning-lab/03_Knowledge/CLI/`, ensure the Folder property has "CLI" as an option.

### 5.1 Add CLI Option to Folder Property
1. Click on Folder property
2. Add option: **"CLI"**
3. Choose color (Blue recommended)
4. Save

---

## ✅ Verification Checklist

Before proceeding to full sync, verify:

- [ ] CLI Architecture category added to Category property
- [ ] Folder property added with CLI option
- [ ] File Path property added (Text)
- [ ] AI Summary property added (Text)
- [ ] Folder Context property added (Text)
- [ ] Related Knowledge property added (Multi-select)
- [ ] AI Category property added (Select)
- [ ] AI Priority property added (Select)
- [ ] AI Relationships property added (Multi-select)
- [ ] Sync Status property added (Select)
- [ ] Last Sync property added (Date)
- [ ] NOTION_API_KEY environment variable set
- [ ] API access test successful

---

## 🚀 Next Steps After Configuration

Once configuration is complete:

1. **Run dry-run sync**:
```bash
cd /z/10_WORKPLACE/Ti/apps/automation
export NOTION_API_KEY=$(cat Z:\00_SECRET\notion.env)
go run . knowledge-sync-ai --path="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\CLI" --dry-run
```

2. **Run full sync**:
```bash
go run . knowledge-sync-ai --path="Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\CLI"
```

3. **Verify in Notion**:
   - Check that 7 CLI docs were created
   - Verify all properties are populated
   - Check that folder context is present
   - Verify AI summary is generated
   - Check that relationships are detected

---

## 📊 Property Reference Table

| Property | Type | Required | CLI-Specific | Purpose |
|----------|------|----------|--------------|---------|
| Tên | Title | ✅ Yes | No | Document name |
| Type | Select | ✅ Yes | No | Document type |
| Status | Select | ✅ Yes | No | Document status |
| Priority | Select | ✅ Yes | No | Priority level |
| Category | Select | ✅ Yes | ✅ Yes | Knowledge category |
| Folder | Select | ✅ New | ✅ Yes | Folder name |
| File Path | Text | ✅ New | ✅ Yes | Full file path |
| AI Summary | Text | ✅ New | ✅ Yes | AI-generated summary |
| Folder Context | Text | ✅ New | ✅ Yes | Folder file list |
| Related Knowledge | Multi-select | ✅ New | ✅ Yes | Related files |
| AI Category | Select | ✅ New | ✅ Yes | AI-determined category |
| AI Priority | Select | ✅ New | ✅ Yes | AI-determined priority |
| AI Relationships | Multi-select | ✅ New | ✅ Yes | AI-detected relations |
| Sync Status | Select | ✅ New | ✅ Yes | Sync tracking |
| Last Sync | Date | ✅ New | ✅ Yes | Sync timestamp |

---

## 🐛 Troubleshooting

### Issue: Property Creation Fails
**Solution**: Ensure you have edit permissions on the database

### Issue: API Access Test Fails
**Solution**: 
- Verify NOTION_API_KEY is correct
- Check that the API key has database access
- Ensure database ID is correct: `34fb60e8155a80c787a6ce3ea9f631af`

### Issue: Sync Creates Pages but Properties Empty
**Solution**: 
- Verify property names match exactly (case-sensitive)
- Check that property types are correct
- Ensure API key has permission to write properties

---

**Status**: Configuration Guide Created ✅
**Next**: Follow this guide to configure Notion database, then run full sync
