---
tags: ["tibrain", "documentation", "provider-claude", "router", "skill"]
scopes: ["providers", "tibrain"]
last_updated: 2026-05-22
---
# Claude-API Pattern Analysis

## Repository: Claude-API
**URL:** https://github.com/KoushikNavuluri/Claude-API
**Location:** Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\Claude-API

## Architecture Patterns

### 1. Cookie-Based Authentication Pattern
**Pattern:** Direct cookie authentication with organization ID extraction
```python
class Client:
    def __init__(self, cookie):
        self.cookie = cookie
        self.organization_id = self.get_organization_id()
    
    def get_organization_id(self):
        url = "https://claude.ai/api/organizations"
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0',
            'Cookie': f'{self.cookie}'
        }
        response = requests.get(url, headers=headers, impersonate="chrome110")
        res = json.loads(response.text)
        uuid = res[0]['uuid']
        return uuid
```

**Key Learnings:**
- Cookie passed during initialization
- Organization ID extracted from first API call
- User-Agent spoofing for browser impersonation
- Chrome 110 impersonation using curl_cffi

### 2. Browser Impersonation Pattern
**Pattern:** curl_cffi for browser fingerprint evasion
```python
from curl_cffi import requests

response = requests.get(url, headers=headers, impersonate="chrome110")
```

**Key Learnings:**
- Use curl_cffi instead of standard requests
- Chrome 110 impersonation for anti-bot evasion
- Maintains browser fingerprint consistency
- Essential for cookie-based authentication

### 3. Standardized Headers Pattern
**Pattern:** Consistent header set across all requests
```python
headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0',
    'Accept-Language': 'en-US,en;q=0.5',
    'Referer': 'https://claude.ai/chats',
    'Content-Type': 'application/json',
    'Sec-Fetch-Dest': 'empty',
    'Sec-Fetch-Mode': 'cors',
    'Sec-Fetch-Site': 'same-origin',
    'Connection': 'keep-alive',
    'Cookie': f'{self.cookie}'
}
```

**Key Learnings:**
- Consistent User-Agent across requests
- Sec-Fetch headers for CORS compliance
- Referer header for origin validation
- Accept-Language for localization

### 4. Streaming Response Pattern
**Pattern:** SSE (Server-Sent Events) parsing
```python
response = requests.post(url, headers=headers, data=payload, impersonate="chrome110", timeout=500)
decoded_data = response.content.decode("utf-8")
decoded_data = re.sub('\n+', '\n', decoded_data).strip()
data_strings = decoded_data.split('\n')
completions = []
for data_string in data_strings:
    json_str = data_string[6:].strip()  # Remove "data: " prefix
    data = json.loads(json_str)
    if 'completion' in data:
        completions.append(data['completion'])
answer = ''.join(completions)
```

**Key Learnings:**
- Handle SSE streaming responses
- Remove "data: " prefix from each line
- Parse JSON from each SSE line
- Concatenate completion chunks

### 5. UUID Generation Pattern
**Pattern:** UUID generation for conversation IDs
```python
def generate_uuid(self):
    random_uuid = uuid.uuid4()
    random_uuid_str = str(random_uuid)
    formatted_uuid = f"{random_uuid_str[0:8]}-{random_uuid_str[9:13]}-{random_uuid_str[14:18]}-{random_uuid_str[19:23]}-{random_uuid_str[24:]}"
    return formatted_uuid
```

**Key Learnings:**
- UUID4 for random UUID generation
- Manual formatting for standard UUID format
- Used for conversation ID generation

### 6. File Upload Pattern
**Pattern:** Multipart file upload with metadata
```python
def upload_attachment(self, file_path):
    url = 'https://claude.ai/api/convert_document'
    file_name = os.path.basename(file_path)
    content_type = self.get_content_type(file_path)
    
    files = {
        'file': (file_name, open(file_path, 'rb'), content_type),
        'orgUuid': (None, self.organization_id)
    }
    
    response = req.post(url, headers=headers, files=files)
    if response.status_code == 200:
        return response.json()
    else:
        return False
```

**Key Learnings:**
- Multipart file upload with organization UUID
- Content type detection based on extension
- Standard requests library for file upload
- JSON response for upload confirmation

### 7. Conversation Management Pattern
**Pattern:** CRUD operations on conversations
```python
def list_all_conversations(self):
    url = f"https://claude.ai/api/organizations/{self.organization_id}/chat_conversations"
    response = requests.get(url, headers=headers, impersonate="chrome110")
    return response.json()

def create_new_chat(self):
    url = f"https://claude.ai/api/organizations/{self.organization_id}/chat_conversations"
    uuid = self.generate_uuid()
    payload = json.dumps({"uuid": uuid, "name": ""})
    response = requests.post(url, headers=headers, data=payload, impersonate="chrome110")
    return response.json()

def delete_conversation(self, conversation_id):
    url = f"https://claude.ai/api/organizations/{self.organization_id}/chat_conversations/{conversation_id}"
    response = requests.delete(url, headers=headers, data=payload, impersonate="chrome110")
    return response.status_code == 204
```

**Key Learnings:**
- RESTful CRUD operations
- Organization-scoped conversation URLs
- UUID-based resource identification
- Status code checking for success

## Implementation Recommendations for Ti Router

### 1. Cookie Authentication
- Implement cookie-based authentication
- Extract organization/session IDs from first call
- Use curl_cffi for browser impersonation

### 2. Browser Impersonation
- Use curl_cffi instead of standard requests
- Implement Chrome 110 impersonation
- Maintain consistent User-Agent

### 3. Standardized Headers
- Create header template function
- Include Sec-Fetch headers for CORS
- Set consistent Referer and Accept-Language

### 4. Streaming Response Handling
- Implement SSE parsing for streaming responses
- Handle "data: " prefix removal
- Concatenate response chunks

### 5. UUID Generation
- Use UUID4 for random ID generation
- Implement standard UUID formatting
- Use for conversation/session IDs

### 6. File Upload
- Implement multipart file upload
- Detect content type from extension
- Include metadata with uploads

### 7. Resource Management
- Implement CRUD operations for resources
- Use organization-scoped URLs
- Check status codes for success

## Priority
High - Apply these patterns for cookie-based authentication and browser impersonation.
