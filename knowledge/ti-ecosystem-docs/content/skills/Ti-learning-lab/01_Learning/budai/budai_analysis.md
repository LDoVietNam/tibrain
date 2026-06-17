# budai Repository Analysis

> **Repository**: https://github.com/manhcuongk55/budai
> **Local Path**: Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\budai
> **Analysis Date**: 2026-04-30
> **Mode**: learn
> **Status**: ✅ Auto-Analyzed

## 📊 Quick Summary

**Tech Stack:**
- Language: Python
- Framework: pip
- Dependencies: 16


## 📈 Repository Statistics

- Total Files: 100
- Total Directories: 31
- Total Lines of Code: 23129
- Largest File: Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\budai\.git\objects\pack\pack-98711ef66fe0196caf233e0e1343adff9416b620.pack (2.33 MB)

## 🔍 Analysis Sections

### 1. Tech Stack
## 📦 Key Dependencies

- fastapi==0.115.6
- uvicorn[standard]==0.34.0
- python-multipart==0.0.20
- pydantic==2.10.4
- pydantic-settings==2.7.1
- openai==1.59.5
- google-genai==1.5.0
- httpx==0.28.1
- chromadb==0.6.3
- langchain-text-splitters==0.3.6
- pypdf==5.1.0
- python-docx==1.1.2
- chardet==5.2.0
- python-dotenv==1.0.1
- aiofiles==24.1.0
- pennylane>=0.39.0

## 📊 Dependency Graph

```mermaid
graph TD
    A[Project]
    B[Core Dependencies]
    A --> B
    B --> C0[fastapi==0.115.6]
    B --> C1[uvicorn[standard]==0.34.0]
    D[AI/ML Dependencies]
    A --> D
    D --> E0[chromadb==0.6.3]
    D --> E1[langchain-text-splitters==0.3.6]
    D --> E2[openai==1.59.5]
    F[HTTP/Web Dependencies]
    A --> F
    F --> G0[httpx==0.28.1]
    H[Other Dependencies]
    A --> H
    H --> I0[aiofiles==24.1.0]
    H --> I1[chardet==5.2.0]
    H --> I2[google-genai==1.5.0]
    H --> I3[pennylane>=0.39.0]
    H --> I4[pydantic-settings==2.7.1]
    H --> I5[pydantic==2.10.4]
    H --> I6[pypdf==5.1.0]
    H --> I7[python-docx==1.1.2]
    H --> I8[python-dotenv==1.0.1]
    H --> I9[python-multipart==0.0.20]
```


### 2. Architecture Analysis
## 📖 README

# budAI 🪷 — The Enlightened AI OS

**budAI** (Hệ Điều Hành AI Bát Nhã) is fundamentally different from traditional AI gateways. It is not just about routing requests; it is an **Ethical AI Operating System** designed to ensure that all AI interactions are rooted in truth, compassion, and non-harm, while maintaining absolute cryptographic privacy between micro-agents.

Inspired by the Heart Sutra (Bát Nhã Tâm Kinh), budAI forces every LLM response to pass through a deep learning Enlightenment Network before reaching the user.

## 🌟 The 4 Pillars of budAI

### 1. Prajna Deep Learning Network (Màng Lọc Bát Nhã)
Every AI response is evaluated by 4 concurrent AI-as-Judge classifiers:
- **Truth (Sự thật):** Reject hallucinations; verify facts against source context.
- **Compassion (Từ bi):** Ensure responses are empathetic and reduce, not increase, suffering.
- **Emptiness (Tính không):** Maintain open-mindedness; multi-perspective answers without dogmatic attachment.
- **Non-Harm (Không gây hại):** Strictly reject toxic, manipulative, or dangerous outputs.
If an answer fails soft checks, budAI forces a **Rewrite Loop**. If it fails the Harm check, it is compassionately **Rejected**.

### 2. Basao Protocol (Giao tiếp Agentic An toàn)
When budAI delegates tasks to other systems (e.g., GoClaw micro-agents), it uses the **Basao Protocol** — *"Hiểu nhau mà không cần biết bên trong"*.
- **Zero Database Exposure:** Agents communicate solely via semantic **Intent Envelopes**. No SQL, no schemas, no connection strings are ever transmitted.
- **Basao Firewall:** Intercepts and blocks any attempt at data exfiltration, SQL injection, or replay attacks between agents.

### 3. ZK Proof of Truth (Bằng chứng ZK)
Inspired by ZK-Credit-Score architectures, budAI employs **Zero-Knowledge Proofs (ZK-SNARKs)**.
When the Prajna Network approves a response, it generates a cryptographic proof (`pi_a`, `pi_b`, `pi_c`). The user can verify mathematically that the AI response genuinely passed the Truth and Compassion thresholds against the private internal context—without budAI ever needing to expose its private knowledge base or internal model weights.

### 4. Quantum-Ready ⚛️ (PennyLane)
budAI is built for the post-quantum era:
- **Quantum Prajna Circuits:** Option to run Truth and Emptiness evaluations using variational quantum circuits on simulators or real quantum hardware (IBM Q, Amazon Braket).
- **Quantum Crypto:** Utilizes true quantum randomness for nonces and simulates BB84 quantum key exchange for unbreakable Basao Protocol security.
- **Quantum Embeddings:** Uses quantum feature maps for potentially exponential speedups in semantic similarity searches.

---

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/manhcuongk55/budai.git
cd budai

# Install dependencies
pip install -r requirements.txt

# Run the API server
uvicorn app.main:app --reload --port 8000
```

## 🛠️ Architecture

budAI acts as a transparent proxy and OS layer between users and foundational LLMs (OpenRouter, Anthropic, Gemini, etc.).

1. **User asks question** -> **Core RAG Pipeline** retrieves context.
2. **Cost Router** selects the most efficient LLM to generate a draft answer.
3. **Prajna Network** intercepts the draft -> Runs 4 parallel classifiers.
4. **Rewrite/Reject** logic is applied if thresholds aren't met.
5. **ZK Prover** creates a Zero-Knowledge Proof.
6. **API returns** the final enlightened answer + ZK validation proof.

## 🤝 Integrations
- **GoClaw:** budAI exposes its capabilities to GoClaw's multi-agent framework via an MCP (Model Context Protocol) Bridge.
- **Berty:** (Coming soon) P2P encrypted chat integration for complete privacy.

---
*Vì nhân loại sự thật.* 🪷


### 3. Purpose & Use Case
[TODO]: Analyze README and documentation to understand primary goal, target users, and key features

### 4. Directory Structure
[TODO]: Document key directories and their purposes

### 5. Strengths
[TODO]: Identify strengths of this repository

### 6. Weaknesses
[TODO]: Identify weaknesses or areas for improvement

### 7. Integration Potential
- Compatibility with Ti CLI: [TODO]
- Integration options: [TODO]
- Complexity: [TODO]

### 8. Recommendations
[TODO]
## 🎨 Detected Design Patterns

### Singleton (80% confidence)
**Category:** Creational
**Description:** Singleton pattern detected - ensures only one instance exists
**Files:** 14

### Factory (70% confidence)
**Category:** Creational
**Description:** Factory pattern detected - object creation logic
**Files:** 17

### Repository (90% confidence)
**Category:** Architectural
**Description:** Repository pattern detected - data access abstraction
**Files:** 1

### Dependency Injection (80% confidence)
**Category:** Architectural
**Description:** Dependency injection detected - IoC container usage
**Files:** 18


## 🏗️ Architecture Analysis

### Architecture Pattern
**Pattern:** Microservices

### Tech Stack Details
**API Style:** REST
**Database:** Unknown
**Build Tools:** pip, Docker
**CI Platform:** None detected

### Key Directories
- **app**: Application code
- **docs**: Documentation
- **tests**: Test suites

### Data Flow
```
Entry point → Response
```

learn%!s(MISSING)%!s(MISSING)## 📝 Notes

**Analysis Mode**: %!s(MISSING)

**Next Steps**:
- [x] Fetch README and documentation
- [x] Analyze source code structure
- [x] Perform architecture mapping
- [x] Detect design patterns
- [ ] Assess integration feasibility

---
*Generated by Ti CLI /repo command with professional analysis*
