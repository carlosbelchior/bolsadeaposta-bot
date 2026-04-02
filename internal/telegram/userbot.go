package telegram

import (
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/queue"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type terminalAuth struct{}

func (terminalAuth) Phone(ctx context.Context) (string, error) {
	fmt.Print("📱 Digite seu número de telefone (com DDI, ex: +5511999999999): ")
	reader := bufio.NewReader(os.Stdin)
	phone, _ := reader.ReadString('\n')
	return strings.TrimSpace(phone), nil
}

func (terminalAuth) Password(ctx context.Context) (string, error) {
	fmt.Print("🔑 Digite sua senha de dupla verificação 2FA (ou aperte Enter se não tiver): ")
	reader := bufio.NewReader(os.Stdin)
	pwd, _ := reader.ReadString('\n')
	return strings.TrimSpace(pwd), nil
}

func (terminalAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("não suportado")
}

func (terminalAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (terminalAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("💬 Digite o código de 5 dígitos recebido no Telegram: ")
	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	return strings.TrimSpace(code), nil
}

// StartUserbot initializes the MTProto client, performs interactive login if needed,
// and starts listening for incoming messages from the target username.
func StartUserbot(ctx context.Context, worker *queue.Worker) error {
	sessionStorage := &telegram.FileSessionStorage{
		Path: filepath.Join(".", "session.json"),
	}

	dispatcher := tg.NewUpdateDispatcher()

	client := telegram.NewClient(config.TelegramAPIID, config.TelegramAPIHash, telegram.Options{
		SessionStorage: sessionStorage,
		UpdateHandler:  dispatcher,
	})

	// Setup message handler
	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		handleIncomingMessage(ctx, u, e, worker, client.API())
		return nil
	})

	dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
		handleIncomingChannelMessage(ctx, u, e, worker, client.API())
		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		// Log in se necessário usando o terminal
		if err := client.Auth().IfNecessary(ctx, auth.NewFlow(
			terminalAuth{},
			auth.SendCodeOptions{},
		)); err != nil {
			return fmt.Errorf("falha na autenticação via terminal: %w", err)
		}

		self, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("falha ao carregar perfil: %w", err)
		}
		log.Printf("🤖 Userbot Logado com Sucesso! Usuário autenticado: @%s (%s)", self.Username, self.Phone)

		log.Println("🎧 Escutando mensagens silenciosamente...")
		<-ctx.Done()
		return ctx.Err()
	})
}

func handleIncomingMessage(ctx context.Context, u *tg.UpdateNewMessage, e tg.Entities, worker *queue.Worker, api *tg.Client) {
	msg, ok := u.Message.(*tg.Message)
	if !ok {
		return // Not a regular text message
	}

	// Verify if sender is the Target Username
	senderID := msg.GetPeerID()
	var username string

	if user, ok := senderID.(*tg.PeerUser); ok {
		if entity, found := e.Users[user.UserID]; found {
			username = entity.Username
		}
	}

	if matchTarget(username) {
		processTipText(msg.Message, worker)
	}
}

func handleIncomingChannelMessage(ctx context.Context, u *tg.UpdateNewChannelMessage, e tg.Entities, worker *queue.Worker, api *tg.Client) {
	msg, ok := u.Message.(*tg.Message)
	if !ok {
		return // Not a regular text message
	}

	senderID := msg.GetPeerID()
	var channelName string

	if channel, ok := senderID.(*tg.PeerChannel); ok {
		if entity, found := e.Channels[channel.ChannelID]; found {
			channelName = entity.Username
		}
	}

	if matchTarget(channelName) {
		processTipText(msg.Message, worker)
	}
}

func matchTarget(username string) bool {
	// Normalize by stripping @
	normTarget := strings.TrimSpace(strings.ReplaceAll(config.TelegramTargetUsername, "@", ""))
	normFound := strings.TrimSpace(strings.ReplaceAll(username, "@", ""))
	
	if normTarget == "" {
		return false // Failsafe
	}
	return strings.EqualFold(normTarget, normFound)
}

func processTipText(text string, worker *queue.Worker) {
	if text == "" {
		return
	}
	
	tip, err := ParseTipMessage(text)
	if err == nil {
		log.Printf("✅ Sinal capturado vindo de @%s! Enviando para o processamento em background...", config.TelegramTargetUsername)
		worker.AddTip(tip)
	}
}
