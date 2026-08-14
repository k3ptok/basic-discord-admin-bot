# 🤖 Custom Event Reminder & Automod Discord Bot

A high-performance Discord bot written in Go using the `discordgo` library. This bot provides advanced, multi-tenant server configurations for scheduled event notifications alongside real-time moderation engines like rolling-window anti-spam limits and file-based domain blocklists.

## ✨ Features
* **Namespaced Slash Commands**: Clean structure nested under `/mybot config` and `/mybot mod` namespaces.
* **Custom Multi-Reminders**: Allows server administrators to set custom timing alerts (e.g., `15m`, `2h`, `1d`) for scheduled server events.
* **Persistent Configuration**: Tracks configuration adjustments safely per server via an asynchronous JSON database system (`config.json`).
* **Automated Account Isolation**: Instantly protects your community by deleting phishing text, matching bad URLs, tracking rate-limits, and applying a 24-hour user timeout.
* **Structured Local Logging**: Captures thread-safe application updates in a clean, JSON-structured local text file (`bot.log`) via `slog`.

---

## 🛠️ Prerequisites

Before installing, ensure you have the following packages installed on your deployment computer:
* **Go Compiler**: Version `1.21` or higher ([Download Go](https://go.dev))
* **Git**: To clone the source repository.

---

## 🚀 Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com/k3ptok/basic-discord-admin-bot
cd basic-discord-admin-bot
```

### 2. Configure Your System Environment Variables
The bot securely draws its active operational token using native system environment bindings.

Find your and save your Discord Server ID and Discord Token when you create and add a bot to your desired Discord Server.
Create a `.env` file and paste both inside in the following format:
```bash
DISCORD_KEY="BOT_TOKEN_HERE"
SERVER_ID="SERVER_ID_HERE"
```
Create a `.gitignore` and add `.env`, `bot.log`, `config.json` to it on newlines so that they are not accidentally shared.

### 4. Build and Run the Bot
Compile your Go binary asset dependencies, and execute:

```bash
# Fetch required package dependency libraries 
go mod tidy

# Run the program compiling all files in the current package directory
go run .

```

---

## ⚙️ Discord Portal Configuration (Crucial Step)

For the bot to see channel contents, run moderation routines, and execute commands, you must enable specific parameters in the [Discord Developer Portal](https://discord.com):

1. **Enable Gateway Intents**: Go to your Application ➔ **Bot** tab. Scroll down to **Privileged Gateway Intents** and enable the **Message Content Intent**.
2. **Bot Permissions**: When generating your Invite URL under **OAuth2 ➔ URL Generator**, select the `bot` and `applications.commands` scopes. Check the following permissions:
   * `Manage Messages` (To drop scam hits)
   * `Timeout Members` / `Moderate Members` (To safely isolate targets)
   * `Send Messages` & `Embed Links` (To post alerts)
3. **Role Hierarchy**: In your Discord server settings, drag your bot's integration role **above** standard member roles, or the timeout engine will throw an access error.

---

## ⌨️ Active Command Grid

All controls are protected by the **Manage Server** permission layer:

| Command | Action Description |
| :--- | :--- |
| `/mybot config set-channel` | Designates the text channel where reminders and alerts are delivered. |
| `/mybot config view` | Renders a summary card showing active settings and timestamps. |
| `/mybot config reminders` | Appends a custom time duration string (e.g., `45m`, `3h`) to the alert list. |
| `/mybot mod reload-domains` | Flushes the memory registry and reloads `banned_domains.txt` on the fly. |

---

## 📁 Project Architecture Map

```text
├── main.go             # Application initialization, gateway engine, and loop bootstrap
├── handlers.go         # InteractionCreate / MessageCreate event handler routing logic
├── storage.go          # Thread-safe JSON state save/load engines (config.json)
├── config.json         # Runtime persistent state document (Auto-generated)
├── banned_domains.txt  # Custom plain-text bad domain string list configuration
└── bot.log             # Production structured JSON telemetry log file (Auto-generated)
```

