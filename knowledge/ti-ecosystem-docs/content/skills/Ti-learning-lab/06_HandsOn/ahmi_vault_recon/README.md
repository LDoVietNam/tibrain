
# Ahmi Vault Recon - Automation Suite

## Overview
This is a dedicated "All-in-One" research project for analyzing and automating the Ahmi Vault platform.

## Structure
- **backend/**: Contains the automation logic.
    - `factory/`: Account generator scripts (`gen_account.cjs`, `scheduler.js`).
    - `mailhub/`: Control server and Outlook integration.
- **ui/**: React-based Dashboard for controlling the bot.

## Usage
1.  **Start Backend**:
    ```bash
    cd backend/mailhub
    node server.js
    ```
2.  **Start UI**:
    ```bash
    cd ui
    npm install
    npm run dev
    ```
3.  **Access**: `http://localhost:5173`

## Artifacts
- Accounts: `Z:\knowledge_base\Database\vault_accounts.json`
- Config: `Z:\knowledge_base\Database\mailhub_config.json`
