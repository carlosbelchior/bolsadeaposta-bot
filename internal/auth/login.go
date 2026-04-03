package auth

import (
	"bolsadeaposta-bot/internal/config"
	"fmt"
	"log"
	"strings"
	"os"
	"io"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"golang.org/x/term"
)

var (
	// InputReader is used for terminal interaction. Can be mocked in tests.
	InputReader io.Reader = os.Stdin
	// ConsoleWriter is used for terminal logs.
	ConsoleWriter io.Writer = os.Stdout
)

func readString() (string, error) {
	var input string
	_, err := fmt.Fscanln(InputReader, &input)
	return input, err
}

func readPassword() (string, error) {
	if f, ok := InputReader.(*os.File); ok {
		bytePassword, err := term.ReadPassword(int(f.Fd()))
		return string(bytePassword), err
	}
	// Fallback for non-file readers (like in tests)
	return readString()
}

func LoginFlow(page *rod.Page) error {
	log.Println("⏳ Verificando necessidade de login...")

	// 1. Verifica se já está logado
	userActionEl, err := page.Timeout(config.DelayNavigation).Element(config.SelectorUserActions)
	if err == nil {
		username, _ := userActionEl.Text()
		log.Printf("✅ Já logado como: %s", strings.TrimSpace(username))
		return nil
	}

	// 2. Verifica se está na tela de login
	_, err = page.Timeout(config.DelayNavigation).Element(`input[type="password"]`)
	if err == nil {
		log.Println("🔑 Tela de login detectada.")
		return handleLoginData(page)
	}

	// 3. Fallback
	log.Println("ℹ️ Não foi possível identificar login. Assumindo que já está logado.")
	return nil
}

func handleLoginData(page *rod.Page) error {
	log.Println("🔐 Iniciando preenchimento de login...")

	user := config.BolsaUsername
	pass := config.BolsaPassword

	if user == "" || pass == "" {
		fmt.Fprint(ConsoleWriter, "Digite seu usuário: ")
		user, _ = readString()

		fmt.Fprint(ConsoleWriter, "Digite sua senha: ")
		pass, _ = readPassword()
		fmt.Fprintln(ConsoleWriter)
	} else {
		log.Println("ℹ️ Credenciais carregadas via variáveis de ambiente.")
	}

	// 🔎 Busca inputs
	userInput, err := page.Element(config.SelectorUsernameInput)
	if err != nil {
		return fmt.Errorf("campo de usuário não encontrado: %w", err)
	}

	passInput, err := page.Element(config.SelectorPasswordInput)
	if err != nil {
		return fmt.Errorf("campo de senha não encontrado: %w", err)
	}

	// ✍️ Preenche
	if err := userInput.Input(user); err != nil {
		return fmt.Errorf("erro ao preencher usuário: %w", err)
	}
	if err := passInput.Input(pass); err != nil {
		return fmt.Errorf("erro ao preencher senha: %w", err)
	}

	// ⏎ Envia
	if err := page.Keyboard.Press(input.Enter); err != nil {
		return fmt.Errorf("erro ao enviar login: %w", err)
	}

	// ⏳ Aguarda pós-login
	page.MustWaitLoad()

	// 🔄 Trata possíveis modais
	handleLocationIPModal(page)
	handleDeviceValidationModal(page)

	log.Println("✅ Login processado.")
	return nil
}

func handleLocationIPModal(page *rod.Page) {
	log.Println("⏳ Verificando modal de localização...")

	xpath := `//p[contains(@class, 'link-text') and contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'continue using location via ip address')]`
	modalEl, err := page.Timeout(config.TimeoutModal).ElementX(xpath)
	if err != nil {
		log.Println("ℹ️ Modal de localização não apareceu.")
		return
	}

	log.Println("📍 Modal encontrado. Clicando...")
	_ = modalEl.Click(proto.InputMouseButtonLeft, 1)

	page.MustWaitLoad()
}

func handleDeviceValidationModal(page *rod.Page) {
	log.Println("⏳ Verificando validação de novo dispositivo...")

	xpath := `//div[contains(@class, 'mat-mdc-dialog-title') and contains(text(), 'New device detected')]`
	_, err := page.Timeout(config.TimeoutModal).ElementX(xpath)
	if err != nil {
		log.Println("ℹ️ Modal de validação não apareceu.")
		return
	}

	log.Println("📱 Modal de validação encontrado.")

	fmt.Fprint(ConsoleWriter, "Digite o código enviado para seu e-mail: ")
	code, _ := readString()

	inputs, err := page.Elements(`code-input input`)
	if err != nil || len(inputs) == 0 {
		log.Println("⚠️ Inputs de código não encontrados.")
		return
	}

	for i, char := range code {
		if i < len(inputs) {
			_ = inputs[i].Input(string(char))
		}
	}

	loginBtn, err := page.Element(`mat-dialog-actions button.btn--color`)
	if err == nil {
		_ = loginBtn.Click(proto.InputMouseButtonLeft, 1)
		log.Println("✅ Código enviado com sucesso.")
		page.MustWaitLoad()
	}
}
