package browser

import (
	"bolsadeaposta-bot/internal/auth"
	"bolsadeaposta-bot/internal/config"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func StartBrowserAndAccessSportsbook() (*rod.Browser, *rod.Page, error) {
	url, err := launcher.New().
		Headless(false).
		UserDataDir("./user-data").
		Set("lang", "pt-BR").
		Set("accept-lang", "pt-BR,pt;q=0.9").
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

	log.Println("⏳ Carregando página...")
	if err := page.WaitLoad(); err != nil {
		log.Printf("⚠️ Erro ao esperar carregamento (prosseguindo mesmo assim): %v", err)
	}

	handleAgeModal(page)
	wakeUpPage(page)
	handleCookiesModal(page)
	HandleRealityCheck(page)

	if err := auth.LoginFlow(page); err != nil {
		return nil, nil, fmt.Errorf("erro no fluxo de login: %w", err)
	}

	return browser, page, nil
}

// CheckAndDismissPopups can be called periodically to clear blocking modals
func CheckAndDismissPopups(page *rod.Page) {
	HandleRealityCheck(page)
}

func HandleRealityCheck(page *rod.Page) {
	// O modal pode demorar um pouco para renderizar o botão ou estar em animação
	btn, err := page.Timeout(1 * time.Second).Element(config.SelectorRealityCheckBtn)
	if err == nil {
		log.Println("🕒 Modal de Verificação de Realidade detectado. Clicando em 'Continuar jogando'...")
		_ = btn.ScrollIntoView()
		_ = btn.Click(proto.InputMouseButtonLeft, 1)
		
		// Aguarda sumir para não interferir em cliques subsequentes
		_ = btn.WaitInvisible()
	}
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
	log.Println("⏳ Verificando modal de idade...")
	start := time.Now()
	for time.Since(start) < config.TimeoutModal {
		buttons, err := page.Elements(config.SelectorAgeYesBtn)
		if err == nil {
			for _, btn := range buttons {
				text, _ := btn.Text()
				if strings.TrimSpace(strings.ToLower(text)) == "yes" {
					log.Println("🔞 Modal de idade detectado. Clicando...")
					_ = btn.ScrollIntoView()
					_ = btn.Click(proto.InputMouseButtonLeft, 1)
					page.MustWaitLoad()
					return
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Println("ℹ️ Modal de idade não encontrado.")
}

func handleCookiesModal(page *rod.Page) {
	log.Println("⏳ Verificando modal de cookies...")
	start := time.Now()
	for time.Since(start) < config.TimeoutModal {
		buttons, err := page.Elements(config.SelectorCookieBtn)
		if err == nil {
			for _, btn := range buttons {
				text, _ := btn.Text()
				if strings.Contains(strings.ToLower(text), "cookies") {
					log.Println("🍪 Modal de cookies detectado. Clicando...")
					_ = btn.ScrollIntoView()
					_ = btn.Click(proto.InputMouseButtonLeft, 1)
					page.MustWaitLoad()
					return
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Println("ℹ️ Modal de cookies não encontrado.")
}
