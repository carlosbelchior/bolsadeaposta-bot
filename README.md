# 🤖 Bolsadeaposta Bot (Telegram Userbot + E-Soccer Automation)

Profissional Go-based automation tool designed to passively monitor Telegram signals (tips) from third-party bots and automatically execute bets on the Bolsadeaposta platform. Built with **MTProto** ([gotd/td](https://github.com/gotd/td)) to clone your human Telegram session and [go-rod](https://github.com/go-rod/rod) for dynamic multi-tab browser automation.

---

## 🔥 Key Features

1.  **👻 Ghost Userbot (MTProto)**
    *   Acts as a true Telegram Client logged into your personal account.
    *   Silently intercepts incoming tips from a specifically defined `@TargetBot` in your direct messages or channels.
    *   No manual forwarding needed. Fully automated routing of signals based on Regex matching.

2.  **🌐 Multi-Tab Asynchronous Execution (`queue.Worker`)**
    *   No bottleneck: Capable of receiving dozens of simultaneous betting tips.
    *   Every new tip instantly spawns a dedicated, isolated invisible browser tab.
    *   Each tab handles its own 10-minute validity cycle without delaying other active bets.

3.  **🕵️ Dynamic Live Validations (`crawler` & `betting`)**
    *   Automatically locates the target Event based on the player names from the tip.
    *   **Live Score Validation**: Continually verifies if the current match score still matches the exact score recommended in the Tip. Drops the bet if a goal occurs before the target odds are reached.
    *   **Minimum Odd Evaluation**: Keeps scraping the page every 2 seconds until the Goal Market line (e.g., `Over 4.5`) respects the minimum profitable Odd.

4.  **🛡️ Intelligent Navigation**
    *   Bypasses +18 modals and Cookies automatically.
    *   Retains login session internally minimizing re-auth prompts on site.

---

## 📂 Project Architecture

*   **`internal/config`**: Environment settings extracted from `.env` covering MTProto credentials and Bolsadeaposta scraping rules.
*   **`internal/telegram`**: Houses the strict MTProto Userbot client, interactive terminal authentication (`terminalAuth`), and the Regex `parser` for the tip strings.
*   **`internal/queue`**: The core concurrency engine. Handles pending tips, creates new browser contexts, and orchestrates the 2-second polling cycles.
*   **`internal/crawler`**: Automation methods to navigate through iframes, find matches natively, and parse DOM elements (scores, time).
*   **`internal/betting`**: Executes the exact interactions to place bets onto the betslip evaluating limits and active lines.
*   **`internal/models`**: Centralized structure for elements (`Tip`, `Status`).

---

## 🛠️ Prerequisites

*   [Go](https://go.dev/dl/) (version 1.20+ recommended)
*   Google Chrome or Chromium installed on your system.
*   Registered Telegram `API_ID` and `API_HASH` from [my.telegram.org](https://my.telegram.org).

---

## 🚀 Getting Started

### 1. Clone & Configure
```bash
# Clone the repository
git clone <repo-url>
cd bolsadeaposta-bot

# Install dependencies (gotd + go-rod + godotenv)
go mod tidy
```

### 2. Environment Variables (`.env`)
Copy the `.env.example` file and create a `.env`:
```env
TELEGRAM_API_ID=your_api_id_from_my_telegram_org
TELEGRAM_API_HASH=your_api_hash_from_my_telegram_org
TELEGRAM_TARGET_USERNAME=@usernameOfTipsterBot
```

### 3. Execution & Login
Run the main entry point:
```bash
go run main.go
```
*   **First Run**: Look at your terminal! The application will ask for your Phone Number (e.g., `+5511999999999`) and then prompt you for the standard 5-digit login code sent by the official Telegram App to establish the `session.json`.
*   **Subsequent Runs**: It will skip authentication smoothly and wait silently for targets.

---

## ⚠️ Important Notes

*   **User Data Persistence:** The `./user-data` directory stores your browser session (cookies/cache). The `./session.json` handles your encrypted Telegram session. Do not commit these files to version control!
*   **Liability:** This tool is for **educational and simulation purposes only**. Always ensure you follow the platform's terms of service and gamble responsibly.
