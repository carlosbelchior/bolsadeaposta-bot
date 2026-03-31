package auth

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

func LoginFlow(page *rod.Page) {
	fmt.Println("⏳ Verificando necessidade de login...")

	// 🔥 1. Verifica se já está logado
	el, errLogged := page.Timeout(3 * time.Second).Element(`.useractions .user-name strong`)
	if errLogged == nil {
		username := el.MustText()
		fmt.Printf("✅ Já logado como: %s\n", username)
		return
	}

	// 🔑 2. Verifica se está na tela de login
	_, errInput := page.Timeout(3 * time.Second).Element(`input[type="password"]`)
	if errInput == nil {
		fmt.Println("🔑 Tela de login detectada.")
		handleLoginData(page)
		return
	}

	// 🤷 3. Fallback
	fmt.Println("ℹ️ Não foi possível identificar login. Assumindo que já está logado.")
}

func handleLoginData(page *rod.Page) {
	fmt.Println("🔐 Iniciando preenchimento de login...")

	var user, pass string

	fmt.Print("Digite seu usuário: ")
	fmt.Scanln(&user)

	fmt.Print("Digite sua senha: ")
	fmt.Scanln(&pass)

	// 🔎 Busca inputs
	userInput, err := page.Element(`input[placeholder="Username"]`)
	if err != nil {
		userInput = page.MustElement(`input[type="text"], input[type="email"], input[placeholder*="Usuário"], input[placeholder*="Email"]`)
	}

	passInput, err := page.Element(`input[placeholder="Password"]`)
	if err != nil {
		passInput = page.MustElement(`input[type="password"]`)
	}

	// ✍️ Preenche
	userInput.MustInput(user)
	passInput.MustInput(pass)

	// ⏎ Envia
	page.Keyboard.Press(input.Enter)

	// ⏳ Aguarda pós-login (melhor que WaitLoad)
	time.Sleep(3 * time.Second)

	// 🔄 Trata possíveis modais
	handleLocationIPModal(page)
	handleDeviceValidationModal(page)

	fmt.Println("✅ Login processado.")
}

func handleLocationIPModal(page *rod.Page) {
	fmt.Println("⏳ Verificando modal de localização...")

	el, err := page.Timeout(5 * time.Second).ElementX(`//p[contains(@class, 'link-text') and contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'continue using location via ip address')]`)
	if err != nil {
		fmt.Println("ℹ️ Modal não apareceu.")
		return
	}

	fmt.Println("📍 Modal encontrado. Clicando...")
	el.MustClick()

	time.Sleep(2 * time.Second)
}

func handleDeviceValidationModal(page *rod.Page) {
	fmt.Println("⏳ Verificando validação de novo dispositivo...")

	_, err := page.Timeout(15 * time.Second).ElementX(`//div[contains(@class, 'mat-mdc-dialog-title') and contains(text(), 'New device detected')]`)
	if err != nil {
		fmt.Println("ℹ️ Modal de validação não apareceu.")
		return
	}

	fmt.Println("📱 Modal de validação encontrado.")

	var code string
	fmt.Print("Digite o código enviado para seu e-mail: ")
	fmt.Scanln(&code)

	inputs, _ := page.Elements(`code-input input`)

	if len(inputs) > 0 {
		for i, char := range code {
			if i < len(inputs) {
				inputs[i].MustInput(string(char))
			}
		}

		time.Sleep(1 * time.Second)

		loginBtn, err := page.Element(`mat-dialog-actions button.btn--color`)
		if err == nil {
			loginBtn.MustClick()
			fmt.Println("✅ Código enviado com sucesso.")
			time.Sleep(3 * time.Second)
		}
	}
}
