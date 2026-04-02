package browser

import (
	"bolsadeaposta-bot/internal/auth"
	"bolsadeaposta-bot/internal/config"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func LoadPageFlow() (*rod.Browser, *rod.Page, error) {
	url, err := launcher.New().
		Headless(false).
		UserDataDir("./user-data").
		Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao lançar navegador: %w", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, nil, fmt.Errorf("erro ao conectar ao navegador: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: config.BaseURL})
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao abrir página: %w", err)
	}

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             1920,
		Height:            1080,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}); err != nil {
		return nil, nil, fmt.Errorf("erro ao configurar viewport: %w", err)
	}

	fmt.Println("⏳ Carregando página...")
	if err := page.WaitLoad(); err != nil {
		fmt.Printf("⚠️ Erro ao esperar carregamento (prosseguindo mesmo assim): %v\n", err)
	}

	handleAgeModal(page)
	wakeUpPage(page)
	handleCookiesModal(page)

	if err := auth.LoginFlow(page); err != nil {
		return nil, nil, fmt.Errorf("erro no fluxo de login: %w", err)
	}

	return browser, page, nil
}

func wakeUpPage(page *rod.Page) {
	for i := 0; i < 5; i++ {
		_, _ = page.Eval(`() => {
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
	fmt.Println("⏳ Verificando modal de idade...")
	start := time.Now()
	for time.Since(start) < config.TimeoutModal {
		buttons, err := page.Elements(config.SelectorAgeYesBtn)
		if err == nil {
			for _, btn := range buttons {
				text, _ := btn.Text()
				if strings.TrimSpace(strings.ToLower(text)) == "yes" {
					fmt.Println("🔞 Modal de idade detectado. Clicando...")
					_ = btn.ScrollIntoView()
					_ = btn.Click(proto.InputMouseButtonLeft, 1)
					page.MustWaitLoad()
					return
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("ℹ️ Modal de idade não encontrado.")
}

func handleCookiesModal(page *rod.Page) {
	fmt.Println("⏳ Verificando modal de cookies...")
	start := time.Now()
	for time.Since(start) < config.TimeoutModal {
		buttons, err := page.Elements(config.SelectorCookieBtn)
		if err == nil {
			for _, btn := range buttons {
				text, _ := btn.Text()
				if strings.Contains(strings.ToLower(text), "cookies") {
					fmt.Println("🍪 Modal de cookies detectado. Clicando...")
					_ = btn.ScrollIntoView()
					_ = btn.Click(proto.InputMouseButtonLeft, 1)
					page.MustWaitLoad()
					return
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("ℹ️ Modal de cookies não encontrado.")
}
