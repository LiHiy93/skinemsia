package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"skinemsia/internal/store"
)

type Bot struct {
	api       *tgbotapi.BotAPI
	store     *store.Store
	webAppURL string
}

func New(token string, st *store.Store, webAppURL string) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("BOT_TOKEN is empty")
	}
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	log.Printf("Bot authorized: @%s", api.Self.UserName)
	return &Bot{api: api, store: st, webAppURL: webAppURL}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		go b.handleMessage(update.Message)
	}
}

func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	ctx := context.Background()
	log.Printf("Bot message from %d (@%s): %q", msg.From.ID, msg.From.UserName, msg.Text)

	username := msg.From.UserName
	_, _ = b.store.UpsertUser(ctx, msg.From.ID, username, msg.From.FirstName, msg.From.LastName)

	text := strings.TrimSpace(msg.Text)
	switch {
	case text == "/start" || strings.HasPrefix(text, "/start "):
		b.handleStart(msg)
	case text == "/help":
		b.handleHelp(msg)
	case text == "/events":
		b.handleEvents(msg, ctx)
	case text == "/new":
		b.handleNew(msg)
	case strings.HasPrefix(text, "/join "):
		code := strings.TrimPrefix(text, "/join ")
		b.handleJoin(msg, ctx, strings.TrimSpace(code))
	default:
		b.send(msg.Chat.ID, "Используй команды:\n/start — меню\n/events — мои события\n/help — помощь")
	}
}

// sendWithURL sends a message with an optional URL button.
// If URL is not HTTPS (e.g. localhost), sends text only — Telegram rejects http:// buttons.
func (b *Bot) sendWithURL(chatID int64, text, btnText, url string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if strings.HasPrefix(url, "https://") {
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(btnText, url),
			),
		)
		msg.ReplyMarkup = kb
	}
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Bot send error: %v", err)
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)
	if strings.HasPrefix(text, "/start ") {
		code := strings.TrimPrefix(text, "/start ")
		b.handleJoin(msg, context.Background(), strings.TrimSpace(code))
		return
	}

	name := msg.From.FirstName
	welcome := fmt.Sprintf(
		"Привет, %s! 👋\n\n*Скинемся* — приложение для честного разделения расходов.\n\nОткрой приложение и создай своё первое событие:",
		name,
	)
	b.sendWithURL(msg.Chat.ID, welcome, "🚀 Открыть Скинемся", b.webAppURL)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	text := `*Скинемся* — разделение расходов

*Команды:*
/start — главное меню
/events — список моих событий
/new — создать новое событие
/join КОД — войти по коду приглашения
/help — эта справка

*Как пользоваться:*
1. Открой приложение через кнопку меню
2. Создай событие (пикник, поездка и т.д.)
3. Поделись кодом с друзьями
4. Добавляйте расходы и выбирайте, на кого делить
5. Все видят свою сумму и номер телефона сборщика`

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "Markdown"
	if _, err := b.api.Send(reply); err != nil {
		log.Printf("Bot send error (help): %v", err)
	}
}

func (b *Bot) handleEvents(msg *tgbotapi.Message, ctx context.Context) {
	events, err := b.store.ListUserEvents(ctx, msg.From.ID)
	if err != nil || len(events) == 0 {
		b.sendWithURL(msg.Chat.ID, "У тебя пока нет событий. Создай первое!", "➕ Создать событие", b.webAppURL)
		return
	}

	var sb strings.Builder
	sb.WriteString("*Твои события:*\n\n")
	for _, e := range events {
		icon := "🟢"
		if e.Status == "archived" {
			icon = "📦"
		}
		sb.WriteString(fmt.Sprintf("%s *%s*\nКод: `%s`\n\n", icon, e.Title, e.JoinCode))
	}

	b.sendWithURL(msg.Chat.ID, sb.String(), "📱 Открыть приложение", b.webAppURL)
}

func (b *Bot) handleNew(msg *tgbotapi.Message) {
	b.sendWithURL(msg.Chat.ID, "Открой приложение и создай новое событие:", "➕ Создать событие", b.webAppURL+"#/create")
}

func (b *Bot) handleJoin(msg *tgbotapi.Message, ctx context.Context, code string) {
	if code == "" {
		b.send(msg.Chat.ID, "Укажи код события: /join КОД")
		return
	}

	userID := msg.From.ID
	_, _ = b.store.UpsertUser(ctx, userID, msg.From.UserName, msg.From.FirstName, msg.From.LastName)

	u, err := b.store.GetUserByTelegramID(ctx, userID)
	if err != nil {
		b.send(msg.Chat.ID, "Произошла ошибка. Попробуй ещё раз.")
		return
	}

	e, err := b.store.JoinEvent(ctx, strings.ToUpper(code), u.ID)
	if err != nil {
		b.send(msg.Chat.ID, "Событие не найдено. Проверь код приглашения.")
		return
	}

	joinURL := fmt.Sprintf("%s#/event/%d", b.webAppURL, e.ID)
	b.sendWithURL(msg.Chat.ID, fmt.Sprintf("✅ Ты присоединился к событию *%s*!", e.Title), "📱 Открыть событие", joinURL)
}

func (b *Bot) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Bot send error: %v", err)
	}
}
