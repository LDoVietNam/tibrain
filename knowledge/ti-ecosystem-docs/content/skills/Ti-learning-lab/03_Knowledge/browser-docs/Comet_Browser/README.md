# Comet Browser Research

## Research Summary
Research confirms that **Comet Browser** is an AI-powered browser developed by **Perplexity AI**, focused on "agentic" capabilities (acting on behalf of the user). However, it has been the subject of significant security controversy due to a hidden **MCP (Model Context Protocol) API**.

Key findings indicate that Comet Browser contains a hidden, undocumented API (`chrome.perplexity.mcp.addStdioServer`) that grants embedded extensions (specifically the "Agentic" extension) the ability to execute arbitrary local commands and launch applications without user consent. This vulnerability was exposed by security firm **SquareX**, highlighting risks of "device takeover" and "hidden IT".

## Key Findings

1.  **Hidden MCP API Vulnerability**:
    *   Comet Browser implements a custom MCP API (`chrome.perplexity.mcp.addStdioServer`) that bypasses standard browser sandboxing.
    *   This API allows embedded extensions (like the "Agentic" extension) to execute arbitrary code on the host machine.
    *   The API can be triggered directly from the `perplexity.ai` website, creating a covert channel for potential attacks.

2.  **Embedded & Hidden Extensions**:
    *   Comet comes pre-loaded with two hidden extensions: **Analytics** and **Agentic**.
    *   These extensions are hidden from the user's extension dashboard, making them impossible to disable or monitor.
    *   The "Agentic" extension has persistent access to the MCP API.

3.  **Security Risks**:
    *   **Arbitrary Code Execution**: Attackers could exploit this via "extension stomping" (spoofing the ID of the embedded extension), XSS, or Man-in-the-Middle attacks.
    *   **Device Takeover**: SquareX demonstrated a proof-of-concept where they launched ransomware (WannaCry) via this vulnerability.
    *   **Lack of Documentation**: The API is largely undocumented, leaving security teams blind to the risks.

4.  **Perplexity's Response**:
    *   Perplexity has reportedly "quietly fixed" the specific vulnerability demonstrated by SquareX (error message "Local MCP is not enabled" now appears).
    *   Perplexity positions Comet as an "AI assistant browser" with defense-in-depth security (classifiers, prompt injection mitigation), but the MCP API implementation contradicts standard browser security principles.

## Detailed Analysis

### Architecture & Agentic Capabilities
Comet is designed to be an "action-oriented" browser. Its "Agentic" extension is the core component that enables the AI to perform tasks on the user's behalf. The **MCP API** was likely intended to facilitate this by allowing the browser agent to interact with local tools and applications (a key feature of the Model Context Protocol). However, the implementation lacked the necessary permission boundaries (registry entries, user consent) found in traditional native messaging APIs.

### The MCP API (`chrome.perplexity.mcp.addStdioServer`)
This API is the critical mechanism. It allows the browser to spawn local processes (stdio servers), which is essential for MCP but dangerous if unchecked. In standard MCP implementations (like Claude Desktop), the user must explicitly configure which servers run. In Comet, this was seemingly open to the embedded extensions by default.

### Security Implications for Enterprise
For enterprises, Comet Browser represents a "Shadow IT" risk. Because it behaves like a privileged agent rather than a sandboxed browser, it introduces a new attack surface. Security analysts recommend treating AI browsers like Comet as "unmanaged apps" or "unsanctioned applications" until they provide transparency and control (e.g., ability to disable embedded extensions).

## Forensic Analysis (Local)

**Path**: `C:\Users\MIN\AppData\Local\Perplexity\Comet\Application`
**Version**: `143.2.7499.37654`
**Extension**: `agents.crx` (Extracted to `Z:\knowledge_base\04_Research\00_Active_Research\Comet_Browser\agents_extension`)

**Code Confirmation**:
In `background.js` (Line ~20265), the extension explicitly calls the private API:
```javascript
handler: ({ name: n, command: r, env: s }) => chrome.perplexity.mcp.addStdioServer(n, r, s)
```
This confirms that the installed version of Comet Browser **has the capability** to spawn arbitrary local processes via this extension.

## Mitigation & Solutions

### 1. Neutralization (Disable Agentic Extension)
To safely use Comet Browser without the risk of the hidden agent executing commands:
*   **Action**: Rename or delete the `agents.crx` file in the `default_apps` directory.
*   **Path**: `C:\Users\MIN\AppData\Local\Perplexity\Comet\Application\[VERSION]\default_apps\agents.crx`
*   **Result**: The browser will fail to load the hidden extension on startup, disabling the MCP vector while keeping basic browsing functional.

### 2. Isolation (Windows Sandbox)
If full functionality is required for research:
*   Run Comet Browser exclusively inside **Windows Sandbox**.
*   This ensures that any arbitrary code execution is contained within a temporary, disposable environment and cannot access the host file system.

### 3. Strategic Replacement (Spectre Lite Auto)
Instead of relying on a "black box" agent like Comet:
*   **Build**: Use `Browser-Use` + `Playwright` + `Gemini/Claude`.
*   **Control**: We control the code, the API calls, and the permissions.
*   **Transparency**: No hidden extensions or undocumented APIs.

## Sources & Evidence
*   **SquareX Disclosure**: "Hidden API in Comet AI browser raises security red flags for enterprises" (CSO Online, Hackread).
*   **Vulnerability Details**: "Comet Browser Flaw Lets Hidden API Run Commands on Users’ Devices" (Hackread).
*   **Perplexity Blog**: "Mitigating Prompt Injection in Comet" (Discusses their defense architecture, though not directly addressing the MCP flaw in detail).
*   **Local Forensic Analysis**: Confirmed presence of `chrome.perplexity.mcp.addStdioServer` in `agents.crx`.

## Conclusion
Comet Browser is a pioneering but currently risky implementation of an "Agentic Browser". Its deep integration of MCP for local control is powerful but was implemented with a "move fast, break things" approach that compromised security. Organizations should be cautious and monitor its usage.
