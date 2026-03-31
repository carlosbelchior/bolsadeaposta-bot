package browser

import (
	"betstake-webscrap/internal/auth"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func LoadPageFlow() (*rod.Browser, *rod.Page) {
	url := launcher.New().
		Headless(false).
		UserDataDir("./user-data").
		MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	page := browser.MustPage("https://bolsadeaposta.bet.br/fbook/br-pt/spbk?selectedDefaultTab=Live&selectedLiveSport=1")

	page.MustSetViewport(1920, 1080, 1, false)

	fmt.Println("⏳ Carregando página...")
	page.MustWaitLoad()

	handleAgeModal(page)
	wakeUpPage(page)
	time.Sleep(3 * time.Second)
	handleCookiesModal(page)
	time.Sleep(2 * time.Second)

	auth.LoginFlow(page)

	return browser, page
}

func wakeUpPage(page *rod.Page) {
	for range 5 {
		page.MustEval(`() => {
			document.dispatchEvent(new MouseEvent('mousemove', {
				bubbles: true,
				clientX: Math.random() * window.innerWidth,
				clientY: Math.random() * window.innerHeight
			}));
		}`)
		time.Sleep(30 * time.Millisecond)
	}
}

func handleAgeModal(page *rod.Page) {
	start := time.Now()
	for time.Since(start) < 5*time.Second {
		buttons, _ := page.Elements("#cdk-overlay-0 button")
		for _, btn := range buttons {
			text := strings.TrimSpace(strings.ToLower(btn.MustText()))
			if text == "yes" {
				btn.MustScrollIntoView()
				btn.MustWaitVisible()
				btn.MustWaitEnabled()
				btn.MustClick()
				time.Sleep(2 * time.Second)
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func handleCookiesModal(page *rod.Page) {
	start := time.Now()
	for time.Since(start) < 10*time.Second {
		buttons, _ := page.Elements(`div[id^="cdk-overlay"] button`)
		for _, btn := range buttons {
			text := strings.TrimSpace(strings.ToLower(btn.MustText()))
			if strings.Contains(text, "cookies") {
				btn.MustScrollIntoView()
				btn.MustWaitVisible()
				btn.MustWaitEnabled()
				btn.MustClick()
				time.Sleep(2 * time.Second)
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}
