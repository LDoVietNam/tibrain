# Sydney Pattern Analysis

## Repository: sydney.py
**URL:** https://github.com/vsakkas/sydney.py
**Location:** Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\sydney.py

## Architecture Patterns

### 1. Async Context Manager Pattern
**Pattern:** Async context manager for conversation lifecycle
```python
async def __aenter__(self) -> SydneyClient:
    await self.start_conversation()
    return self

async def __aexit__(self, exc_type, exc_value, traceback) -> None:
    await self.close_conversation()
```

**Key Learnings:**
- Async context manager for resource cleanup
- Automatic conversation start on entry
- Automatic conversation close on exit
- Exception handling in context manager

### 2. Session Management with Proxy Support
**Pattern:** HTTP session with proxy and SSL configuration
```python
async def _get_session(self, force_close: bool = False) -> ClientSession:
    cookies = cookies_as_dict(self.bing_cookies) if self.bing_cookies else {}
    
    if self.session and not self.session.closed and force_close:
        await self.session.close()
        self.session = None
    
    if not self.session:
        self.session = ClientSession(
            headers=CREATE_HEADERS,
            cookies=cookies,
            trust_env=self.use_proxy,  # Use HTTP_PROXY and HTTPS_PROXY env vars
            connector=(
                TCPConnector(verify_ssl=False) if self.use_proxy else None
            ),  # Resolve HTTPS issue when proxy support is enabled
        )
    
    return self.session
```

**Key Learnings:**
- Session reuse with force close option
- Proxy support via environment variables
- SSL verification disabled for proxies
- Cookie integration from string

### 3. WebSocket Connection Pattern
**Pattern:** WebSocket connection with protocol handshake
```python
try:
    self.wss_client = await websockets.connect(
        bing_chathub_url, additional_headers=CHATHUB_HEADERS, max_size=None
    )
except TimeoutError:
    raise ConnectionTimeoutException("Failed to connect to Copilot, connection timed out") from None

await self.wss_client.send(as_json({"protocol": "json", "version": 1}))
await self.wss_client.recv()
```

**Key Learnings:**
- WebSocket connection with custom headers
- Protocol handshake (JSON, version 1)
- Timeout exception handling
- Max size set to None for unlimited size

### 4. Options Sets Building Pattern
**Pattern:** Dynamic options sets based on configuration
```python
options_sets = [option.value for option in DefaultOptions]

# Add conversation style option values
options_sets.extend(
    style.strip()
    for style in self.conversation_style_option_sets.value.split(",")
)

# Build option sets based on whether cookies are used or not
if self.bing_cookies:
    options_sets.extend(option.value for option in CookieOptions)

# Build option sets based on whether search is allowed or not
if not search:
    options_sets.extend(option.value for option in NoSearchOptions)

# Build option sets based on whether a non default GPT persona is used or not
if self.persona != GPTPersonaID.COPILOT:
    options_sets.append(PersonaOptions[self.persona.value.upper()].value)
```

**Key Learnings:**
- Dynamic options set building
- Conditional option extension
- Comma-separated style parsing
- Persona-based option selection

### 5. Request Argument Building Pattern
**Pattern:** Complex nested request structure
```python
arguments: dict = {
    "arguments": [
        {
            "source": "cib",
            "optionsSets": options_sets,
            "allowedMessageTypes": [message.value for message in MessageType],
            "sliceIds": [],
            "verbosity": "verbose",
            "scenario": "CopilotMicrosoftCom",
            "plugins": [],
            "conversationHistoryOptionsSets": [
                option.value for option in ConversationHistoryOptionsSets
            ],
            "gptId": self.persona.value,
            "isStartOfSession": self.invocation_id == 0,
            "message": {
                "author": "user",
                "inputMethod": "Keyboard",
                "timestamp": get_iso_timestamp(),
                "text": prompt,
                "messageType": MessageType.CHAT.value,
                "imageUrl": image_url,
                "originalImageUrl": original_image_url,
            },
            "conversationSignature": self.conversation_signature,
            "participant": {
                "id": self.client_id,
            },
            "tone": str(self.conversation_style.value),
            "extraExtensionParameters": {
                "gpt-creator-persona": {"personaId": self.persona.value}
            },
            "spokenTextMode": "None",
            "conversationId": self.conversation_id,
        }
    ],
    "invocationId": str(self.invocation_id),
    "target": "chat",
    "type": 4,
}
```

**Key Learnings:**
- Nested request structure
- Session detection via invocation_id
- Conversation signature inclusion
- Participant ID tracking
- ISO timestamp generation

### 6. Image Upload Pattern
**Pattern:** Two-stage image upload (metadata + multipart)
```python
# Stage 1: Get upload metadata
image_base64 = None
if not check_if_url(attachment):
    with open(attachment, "rb") as file:
        image_base64 = b64encode(file.read())

data = self._build_upload_arguments(attachment, image_base64)

async with session.post(BING_KBLOB_URL, data=data) as response:
    if response.status != 200:
        raise ImageUploadException(f"Failed to upload image, received status: {response.status}")
    
    response_dict = await response.json()
    if not response_dict["blobId"]:
        raise ImageUploadException("Failed to upload image, Copilot rejected uploading it")
    
    if len(response_dict["blobId"]) == 0:
        raise ImageUploadException("Failed to upload image, received empty image info from Copilot")
```

**Key Learnings:**
- Base64 encoding for local files
- URL detection for remote files
- FormData for multipart upload
- Blob ID validation
- Error handling for upload failures

### 7. Streaming Response Pattern
**Pattern:** WebSocket streaming with delimiter parsing
```python
streaming = True
while streaming:
    objects = str(await self.wss_client.recv()).split(DELIMETER)
    for obj in objects:
        if not obj:
            continue
        response = json.loads(obj)
        
        # Handle type 1 messages when streaming is enabled
        if stream and response.get("type") == 1:
            messages = response["arguments"][0].get("messages")
            if not messages:
                continue
            
            # Skip "Searching the web for..." message
            adaptiveCards = messages[0].get("adaptiveCards")
            if adaptiveCards and adaptiveCards[0]["body"][0].get("inlines"):
                continue
            
            if raw:
                yield response, None
            elif citations:
                yield adaptiveCards[0]["body"][0]["text"], None
            else:
                yield messages[0].get("text"), None
        
        # Handle type 2 messages
        elif response.get("type") == 2:
            # Check if reached conversation limit
            if response["item"].get("throttling"):
                self.number_of_messages = response["item"]["throttling"].get("numUserMessagesInConversation", 0)
                self.max_messages = response["item"]["throttling"]["maxNumUserMessagesInConversation"]
                if self.number_of_messages == self.max_messages:
                    raise ConversationLimitException(f"Reached conversation limit of {self.max_messages} messages")
            
            messages = response["item"].get("messages")
            if not messages:
                result_value = response["item"]["result"]["value"]
                if result_value == ResultValue.THROTTLED.value:
                    raise ThrottledRequestException("Request is throttled")
                elif result_value == ResultValue.CAPTCHA_CHALLENGE.value:
                    raise CaptchaChallengeException("Solve CAPTCHA to continue")
                return
            
            # Yield final response
            yield messages[-1]["text"], suggested_responses
            streaming = False
```

**Key Learnings:**
- Delimiter-based message splitting
- Type-based message handling
- Conversation limit detection
- Throttling detection
- Captcha challenge detection
- Adaptive card parsing

### 8. Conversation Creation Pattern
**Pattern:** Conversation initialization with signature extraction
```python
async def start_conversation(self) -> None:
    session = await self._get_session(force_close=True)
    
    async with session.get(BING_CREATE_CONVERSATION_URL) as response:
        if response.status != 200:
            raise CreateConversationException(f"Failed to create conversation, received status: {response.status}")
        
        response_dict = await response.json()
        if response_dict["result"]["value"] != "Success":
            raise CreateConversationException(f"Failed to authenticate, received message: {response_dict['result']['message']}")
        
        self.conversation_id = response_dict["conversationId"]
        self.client_id = response_dict["clientId"]
        self.conversation_signature = response.headers["X-Sydney-Conversationsignature"]
        self.encrypted_conversation_signature = response.headers["X-Sydney-Encryptedconversationsignature"]
        self.invocation_id = 0
```

**Key Learnings:**
- Conversation ID extraction
- Client ID extraction
- Signature extraction from headers
- Encrypted signature for secure access
- Invocation ID initialization

### 9. Streaming Delta Pattern
**Pattern:** Return only new tokens in streaming
```python
previous_response: str | dict = ""
async for response, suggested_responses in self._ask(..., stream=True):
    if raw:
        yield response
    else:
        new_response = response[len(previous_response) :]
        previous_response = response
        if suggestions:
            yield new_response, suggested_responses
        else:
            yield new_response
```

**Key Learnings:**
- Track previous response
- Calculate delta (new tokens)
- Yield only new content
- Maintain response state

## Implementation Recommendations for Ti Router

### 1. Async Context Manager
- Implement async context managers for resource cleanup
- Automatic initialization on entry
- Automatic cleanup on exit
- Exception handling in context

### 2. Session Management
- Implement session reuse with force close
- Add proxy support via environment variables
- Disable SSL verification for proxies
- Integrate cookies from string

### 3. WebSocket Connection
- Implement WebSocket connections with custom headers
- Add protocol handshake
- Handle timeout exceptions
- Set max size appropriately

### 4. Options Building
- Implement dynamic options set building
- Add conditional option extension
- Parse comma-separated configurations
- Support persona-based selection

### 5. Request Building
- Implement nested request structures
- Add session detection
- Include conversation signatures
- Track participant IDs
- Generate ISO timestamps

### 6. Image Upload
- Implement two-stage image upload
- Support Base64 encoding for local files
- Detect URLs for remote files
- Use FormData for multipart
- Validate blob IDs

### 7. Streaming Response
- Implement delimiter-based parsing
- Handle different message types
- Detect conversation limits
- Detect throttling
- Detect captcha challenges
- Parse adaptive cards

### 8. Conversation Management
- Implement conversation initialization
- Extract conversation IDs
- Extract client IDs
- Extract signatures from headers
- Initialize invocation IDs

### 9. Streaming Delta
- Track previous responses
- Calculate deltas
- Yield only new content
- Maintain response state

## Priority
Medium - Apply these patterns for WebSocket streaming and conversation management.
