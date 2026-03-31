# 🤖 Bolsadeaposta Bot (E-Soccer Automation)

Profissional Go-based automation tool designed to scout and simulate bets on **Asian Handicap** markets, specifically for **E-Soccer (GT Leagues)** on the Bolsadeaposta platform. Built with the powerful [go-rod](https://github.com/go-rod/rod) browser automation library.

---

## 🔥 Key Features

1.  **🛡️ Intelligent Navigation (`browser.LoadPageFlow`)**
    *   Initializes a Chrome instance (visible/headless).
    *   Automatically bypasses age-governance (+18) and Cookie consent modals.
    *   Robustly navigates through the platform's multi-layered layout.

2.  **🕵️ Dynamic Scouter (`crawler.FindLeague` & `GetMatches`)**
    *   Implements **Auto-Scrolling** to trigger Lazy Loading until the "GT Leagues" section is found.
    *   Works seamlessly across nested `iframes`.
    *   Extracts real-time match data: **Teams, Live Scores, Match Time, and 1X2 Odds**.

3.  **📈 Market Deep-Dive (Asian Handicap)**
    *   Clicks into individual matches to reveal detailed markets.
    *   Scans the detail panel (with internal scrolling) for the **"1st Half Asian Handicap"** market.
    *   Maps all available **Lines** and **Odds** in real-time.

4.  **💸 Bet Simulation (`betting.PrepareHandicapBet`)**
    *   Interactive CLI for selecting a specifically targeted match and handicap line.
    *   Automatically fills the bet slip with the desired **Handicap Line** and **Stake (Amount)**.
    *   Stops right before the final confirmation, allowing for manual verification.

---

## 📂 Project Architecture

The project is structured following clean coding principles and modularity:

*   **`internal/auth`**: Orchestrates login flows and authentication challenges.
*   **`internal/betting`**: Contains the logic for interacting with the bet slip and stake inputs.
*   **`internal/browser`**: Configures the Rod browser instance and generic navigation helpers.
*   **`internal/config`**: Handles project settings and persistent user data (`/user-data`).
*   **`internal/crawler`**: Core scraping engine for discovering leagues, matches, and market lines.
*   **`internal/models`**: Defines data structures like `Match` and `HandicapLine`.

---

## 🛠️ Prerequisites

*   [Go](https://go.dev/dl/) (version 1.20+ recommended)
*   Google Chrome or Chromium installed on your system.

---

## 🚀 Getting Started

### 1. Clone & Install Dependencies
```bash
# Clone the repository
git clone <repo-url>
cd bolsadeaposta-bot

# Install Go modules
go mod tidy
```

### 2. Execution
Run the main entry point:
```bash
go run main.go
```

### 3. Usage Flow
1.  Enter the names of the two players (as displayed on the site, e.g., "Giggs", "Ronaldinho").
2.  The bot will find the match and display the current **Scores** and **Handicap Odds**.
3.  Type `s` to simulate a bet.
4.  Follow the prompts to select the **Team**, **Line**, and **Stake**.
5.  Watch the browser automatically populate the bet slip!

---

## ⚠️ Important Notes

*   **User Data Persistence:** The `./user-data` directory stores your browser session (cookies/cache). This prevents the need for manual login on every run.
*   **X-Origin Iframes:** The bot is specially designed to handle the platform's complex iframe structure for betting markets.
*   **Liability:** This tool is for **educational and simulation purposes only**. Always ensure you follow the platform's terms of service and gamble responsibly.
