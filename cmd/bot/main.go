package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	bot_pkg "github.com/Heathcliff-third-space/MediaManager/internal/bot"
	"github.com/Heathcliff-third-space/MediaManager/internal/config"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 检查必要配置
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN 环境变量未设置")
	}

	// 创建机器人管理器
	botManager, err := bot_pkg.NewBotManager(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := bot_pkg.RegisterCommands(botManager.Bot); err != nil {
		log.Printf("注册命令失败: %v", err)
	} else {
		log.Println("成功注册 Telegram 命令")
	}

	// 设置更新配置
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botManager.Bot.GetUpdatesChan(u)

	// 处理中断信号以便优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 同时处理来自 Telegram 的更新和系统信号
	for {
		select {
		case update := <-updates:
			if update.Message != nil { // 如果我们收到一条消息
				if !botManager.IsUserAllowed(update.Message.From.ID) {
					log.Printf("拒绝用户 %s (ID: %d) 的访问", update.Message.From.UserName, update.Message.From.ID)
					botManager.SendAccessDeniedMessage(update.Message.Chat.ID)
					continue
				}
				botManager.HandleMessage(update.Message)
			} else if update.CallbackQuery != nil { // 如果我们收到一个回调查询（按钮点击）
				if !botManager.IsUserAllowed(update.CallbackQuery.From.ID) {
					log.Printf("拒绝用户 %s (ID: %d) 的访问", update.CallbackQuery.From.UserName, update.CallbackQuery.From.ID)
					botManager.SendAccessDeniedMessage(update.CallbackQuery.Message.Chat.ID)
					// 响应回调查询，避免按钮loading状态持续太久
					err := answerCallbackQuery(botManager.Bot, update.CallbackQuery.ID, "访问被拒绝")
					if err != nil {
						log.Printf("响应回调查询失败: %v", err)
					}
					continue
				}
				botManager.HandleCallbackQuery(update.CallbackQuery)
			}

		case <-sigChan:
			log.Println("接收到中断信号，正在关闭...")
			return
		}
	}
}

// answerCallbackQuery 响应回调查询并处理错误
func answerCallbackQuery(bot *tgbotapi.BotAPI, callbackQueryID, text string) error {
	callbackConfig := tgbotapi.NewCallback(callbackQueryID, text)
	_, err := bot.Request(callbackConfig)
	if err != nil {
		log.Printf("响应回调查询失败: %v", err)
		return err
	}
	return nil
}