package bot

import (
	"fmt"
	"github.com/Heathcliff-third-space/MediaManager/internal/util"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Heathcliff-third-space/MediaManager/internal/config"
	"github.com/Heathcliff-third-space/MediaManager/internal/models"
	"github.com/Heathcliff-third-space/MediaManager/internal/services"
)

// Manager 机器人管理器
type Manager struct {
	Bot                *tgbotapi.BotAPI
	mediaServerManager *services.MediaServerManager
	allowedUserIDs     map[int64]bool
}

// NewBotManager 创建新的机器人管理器
func NewBotManager(cfg *config.Config) (*Manager, error) {
	// 初始化 Telegram Bot
	telegramBot, err := initializeTelegramBot(cfg)
	if err != nil {
		return nil, fmt.Errorf("无法初始化Telegram Bot: %v", err)
	}

	log.Printf("已授权账户 %s", telegramBot.Self.UserName)

	// 初始化媒体服务器管理器
	mediaServerManager, err := services.NewMediaServerManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("无法初始化媒体服务器管理器: %v", err)
	}

	// 初始化允许的用户ID映射
	allowedUserIDs := make(map[int64]bool)
	for _, id := range cfg.AllowedUserIDs {
		allowedUserIDs[id] = true
	}
	log.Printf("允许访问的用户ID: %v", cfg.AllowedUserIDs)

	return &Manager{
		Bot:                telegramBot,
		mediaServerManager: mediaServerManager,
		allowedUserIDs:     allowedUserIDs,
	}, nil
}

// IsUserAllowed 检查用户是否有权限使用机器人
func (bm *Manager) IsUserAllowed(userID int64) bool {
	// 如果没有设置允许的用户ID，则允许所有用户访问（向后兼容）
	if len(bm.allowedUserIDs) == 0 {
		return true
	}

	// 检查用户是否在允许列表中
	return bm.allowedUserIDs[userID]
}

// SendAccessDeniedMessage 发送访问拒绝消息
func (bm *Manager) SendAccessDeniedMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "🚫 抱歉，您没有权限使用此机器人。")
	err := sendBotMessage(bm.Bot, msg)
	if err != nil {
		log.Printf("发送访问拒绝消息失败: %v", err)
	}
}

// HandleMessage 处理消息
func (bm *Manager) HandleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	// 只响应特定用户的私聊消息（可选安全措施）
	if message.Chat.Type != "private" {
		return
	}

	switch strings.ToLower(message.Text) {
	case "/start", "/help":
		bm.SendMainMenu(message.Chat.ID, 0)
	case "/serverinfo":
		bm.SendServerInfo(message.Chat.ID, 0)
	case "/users":
		bm.SendUsersInfo(message.Chat.ID, 0)
	case "/search":
		bm.PromptForSearchTerm(message.Chat.ID, 0)
	case "/libraries":
		bm.SendLibrariesList(message.Chat.ID, 0)
	case "/mystats":
		bm.SendMyStats(message.Chat.ID, 0)
	default:
		// 检查是否是搜索查询
		log.Printf("检查是否是搜索查询: ReplyToMessage=%v, Text=%s", message.ReplyToMessage, message.Text)
		if message.ReplyToMessage != nil {
			log.Printf("ReplyToMessage Text: %s", message.ReplyToMessage.Text)
			if strings.Contains(message.ReplyToMessage.Text, "请输入您要搜索的媒体名称") {
				log.Printf("识别为搜索请求: %s", message.Text)
				bm.PerformBookSearch(message.Chat.ID, message.Text)
				return
			}
		}

		// 检查是否是搜索关键词（不依赖ReplyToMessage）
		// 如果用户刚刚点击了搜索按钮，我们就认为下一条消息是搜索词
		bm.PerformBookSearch(message.Chat.ID, message.Text)
		return
	}
}

// HandleCallbackQuery 处理回调查询（按钮点击）
func (bm *Manager) HandleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// 响应回调查询，避免按钮loading状态持续太久
	err := answerCallbackQuery(bm.Bot, callback.ID, "")
	if err != nil {
		log.Printf("响应回调查询失败: %v", err)
	}

	switch callback.Data {
	case "main_menu":
		bm.EditMainMenu(callback.Message.Chat.ID, callback.Message.MessageID)
	case "system_info":
		executeWithLoadingStatus(bm.Bot, callback.Message.Chat.ID, callback.Message.MessageID, "📊 正在获取服务器信息，请稍候...", func() {
			bm.EditServerInfo(callback.Message.Chat.ID, callback.Message.MessageID)
		})
	case "search_books":
		bm.PromptForSearchTerm(callback.Message.Chat.ID, callback.Message.MessageID)
	case "users_list":
		executeWithLoadingStatus(bm.Bot, callback.Message.Chat.ID, callback.Message.MessageID, "👥 正在获取用户信息，请稍候...", func() {
			bm.SendUsersInfo(callback.Message.Chat.ID, callback.Message.MessageID)
		})
	case "my_stats":
		executeWithLoadingStatus(bm.Bot, callback.Message.Chat.ID, callback.Message.MessageID, "📈 正在获取个人统计信息，请稍候...", func() {
			bm.SendMyStats(callback.Message.Chat.ID, callback.Message.MessageID)
		})
	case "libraries_list":
		executeWithLoadingStatus(bm.Bot, callback.Message.Chat.ID, callback.Message.MessageID, "📚 正在获取媒体库信息，请稍候...", func() {
			bm.SendLibrariesList(callback.Message.Chat.ID, callback.Message.MessageID)
		})
	case "help":
		bm.EditHelpMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	}
}

// SendMainMenu 发送主菜单
func (bm *Manager) SendMainMenu(chatID int64, messageID int) {
	msg := tgbotapi.NewMessage(chatID, "🎧 *欢迎使用多服务器媒体管理机器人*\n\n请选择您要执行的操作:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = CreateMainMenu()
	err := sendBotMessage(bm.Bot, msg)
	if err != nil {
		log.Printf("发送主菜单消息失败: %v", err)
	}
}

// EditMainMenu 编辑主菜单
func (bm *Manager) EditMainMenu(chatID int64, messageID int) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, "🎧 *欢迎使用多服务器媒体管理机器人*\n\n请选择您要执行的操作:")
	edit.ParseMode = "Markdown"
	menu := CreateMainMenu()
	edit.ReplyMarkup = &menu
	err := editBotMessage(bm.Bot, edit)
	if err != nil {
		log.Printf("编辑主菜单消息失败: %v", err)
	}
}

// SendServerInfo 发送服务器信息
func (bm *Manager) SendServerInfo(chatID int64, messageID int) {
	// 获取所有服务器的信息
	serverInfo, err := bm.mediaServerManager.GetServerInfoAcrossServers()
	if err != nil {
		if messageID > 0 {
			bm.EditMessage(chatID, messageID, "❌ 获取服务器信息失败: "+err.Error())
		} else {
			bm.SendMessage(chatID, "❌ 获取服务器信息失败: "+err.Error())
		}
		return
	}

	var text string
	if len(serverInfo) == 0 {
		text = "📭 没有找到服务器信息"
	} else {
		text = "📊 *服务器信息*:\n\n"
		for serverType, info := range serverInfo {
			text += fmt.Sprintf("*%s 服务器*:\n", strings.Title(string(serverType)))
			text += fmt.Sprintf("🖥 版本: `%s`\n", info.Version)
			text += fmt.Sprintf("🖥 服务器名: `%s`\n", info.Name)
			text += fmt.Sprintf("💻 操作系统: `%s`\n", info.OS)
			text += fmt.Sprintf("⚙️ 架构: `%s`\n", info.Arch)
			text += "\n"
		}
	}

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = "Markdown"
		menu := CreateServerInfoMenu()
		edit.ReplyMarkup = &menu
		err := editBotMessage(bm.Bot, edit)
		if err != nil {
			log.Printf("编辑服务器信息消息失败: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = CreateServerInfoMenu()
		err := sendBotMessage(bm.Bot, msg)
		if err != nil {
			log.Printf("发送服务器信息消息失败: %v", err)
		}
	}
}

// EditServerInfo 编辑服务器信息
func (bm *Manager) EditServerInfo(chatID int64, messageID int) {
	bm.SendServerInfo(chatID, messageID)
}

// SendLibrariesList 发送媒体库列表
func (bm *Manager) SendLibrariesList(chatID int64, messageID int) {
	allServers := bm.mediaServerManager.GetAllServers()
	var text string

	if len(allServers) == 0 {
		text = "📭 没有找到媒体服务器"
	} else {
		text = "📚 *媒体库列表*:\n\n"

		// 使用并行处理获取所有服务器的媒体库
		var mu sync.Mutex
		var wg sync.WaitGroup
		results := make(map[services.MediaServerType][]models.LibraryInfo)
		errors := make(map[services.MediaServerType]error)

		// 使用信号量控制最大并发数
		maxConcurrency := make(chan struct{}, 4)

		for serverType, server := range allServers {
			wg.Add(1)
			go func(st services.MediaServerType, s models.MediaServer) {
				defer wg.Done()
				// 控制并发数
				maxConcurrency <- struct{}{}
				defer func() { <-maxConcurrency }()

				libraries, err := s.GetLibraries()
				if err != nil {
					mu.Lock()
					errors[st] = err
					mu.Unlock()
					return
				}

				mu.Lock()
				results[st] = libraries
				mu.Unlock()
			}(serverType, server)
		}

		wg.Wait()

		// 按照服务器类型顺序输出结果
		serverTypes := bm.mediaServerManager.GetServerTypes()
		for _, serverType := range serverTypes {
			if _, exists := errors[serverType]; exists {
				text += fmt.Sprintf("*%s 服务器*:\n❌ 获取媒体库失败\n\n", strings.Title(string(serverType)))
				continue
			}

			libraries, exists := results[serverType]
			if !exists {
				continue
			}

			text += fmt.Sprintf("*%s 服务器*:\n", strings.Title(string(serverType)))
			if len(libraries) == 0 {
				text += "📭 暂无媒体库\n"
			} else {
				for _, lib := range libraries {
					if lib.ItemCount > 0 {
						text += fmt.Sprintf("%s %s (%d 个项目)\n", util.GetMediaTypeIcon(lib.MediaType), lib.Name, lib.ItemCount)
					} else {
						text += fmt.Sprintf("%s %s\n", util.GetMediaTypeIcon(lib.MediaType), lib.Name)
					}
				}
			}
			text += "\n"
		}
	}

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = "Markdown"
		menu := CreateLibrariesMenu()
		edit.ReplyMarkup = &menu
		err := editBotMessage(bm.Bot, edit)
		if err != nil {
			log.Printf("编辑媒体库列表消息失败: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = CreateLibrariesMenu()
		err := sendBotMessage(bm.Bot, msg)
		if err != nil {
			log.Printf("发送媒体库列表消息失败: %v", err)
		}
	}
}

// EditLibrariesList 编辑媒体库列表
func (bm *Manager) EditLibrariesList(chatID int64, messageID int) {
	bm.SendLibrariesList(chatID, messageID)
}

// PromptForSearchTerm 提示用户输入搜索词
func (bm *Manager) PromptForSearchTerm(chatID int64, messageID int) {
	// 如果已经有消息ID，则编辑现有消息
	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "🔍 请输入您要搜索的媒体名称、作者或其他关键词：")
		menu := CreateSearchMenu()
		edit.ReplyMarkup = &menu
		err := editBotMessage(bm.Bot, edit)
		if err != nil {
			log.Printf("编辑搜索提示消息失败: %v", err)
		}
	} else {
		// 否则发送新消息
		msg := tgbotapi.NewMessage(chatID, "🔍 请输入您要搜索的媒体名称、作者或其他关键词：")
		menu := CreateSearchMenu()
		msg.ReplyMarkup = &menu
		err := sendBotMessage(bm.Bot, msg)
		if err != nil {
			log.Printf("发送搜索提示消息失败: %v", err)
		}
	}
}

// PerformBookSearch 执行图书搜索
func (bm *Manager) PerformBookSearch(chatID int64, searchTerm string) {
	// 添加调试日志
	log.Printf("执行媒体搜索: %s", searchTerm)

	// 在所有服务器中搜索
	searchResults, err := bm.mediaServerManager.SearchAcrossServers(searchTerm)
	if err != nil {
		log.Printf("搜索出错: %v", err)
		response := fmt.Sprintf("❌ 搜索出错: %v", err)
		msg := tgbotapi.NewMessage(chatID, response)
		msg.ReplyMarkup = CreateMainMenu()
		err := sendBotMessage(bm.Bot, msg)
		if err != nil {
			log.Printf("发送搜索错误消息失败: %v", err)
		}
		return
	}

	// 格式化搜索结果
	response := bm.FormatSearchResults(searchTerm, searchResults)

	// 发送或编辑消息
	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = CreateMainMenu()
	err = sendBotMessage(bm.Bot, msg)
	if err != nil {
		log.Printf("发送搜索结果消息失败: %v", err)
	}
}

// FormatSearchResults 格式化搜索结果
func (bm *Manager) FormatSearchResults(searchTerm string, searchResults map[services.MediaServerType][]models.SearchResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔎 搜索 \"%s\" 的结果:\n\n", searchTerm))

	if len(searchResults) == 0 {
		sb.WriteString("未找到相关媒体。\n")
		return sb.String()
	}

	for serverType, results := range searchResults {
		if len(results) > 0 {
			sb.WriteString(fmt.Sprintf("*%s 服务器*:\n", strings.Title(string(serverType))))
			for i, result := range results {
				if i >= 5 { // 限制每个服务器显示前5个结果
					sb.WriteString(fmt.Sprintf("\n+ 还有 %d 个更多结果...\n", len(results)-5))
					break
				}
				sb.WriteString(fmt.Sprintf("• **%s**\n", result.Title))
				// 根据媒体类型添加图标
				mediaTypeIcon := util.GetMediaTypeIcon(result.Type)
				sb.WriteString(fmt.Sprintf("  %s 类型: %s\n", mediaTypeIcon, result.Type))
				sb.WriteString(fmt.Sprintf("  📁 媒体库: %s\n", result.Library))
				// 添加年份信息
				if result.Year > 0 {
					sb.WriteString(fmt.Sprintf("  📅 年份: %d\n", result.Year))
				}
				// 添加分类信息
				if len(result.Genres) > 0 {
					sb.WriteString(fmt.Sprintf("  🏷️ 分类: %s\n", strings.Join(result.Genres, ", ")))
				}
				// 添加概述信息
				if result.Overview != "" {
					sb.WriteString(fmt.Sprintf("  📝 概述: %s\n", result.Overview))
				}
				// 添加大小信息
				if result.Size > 0 {
					sb.WriteString(fmt.Sprintf("  💾 大小: %s\n", util.FormatBytes(result.Size)))
				}
				// 添加添加时间信息
				if result.AddedAt > 0 {
					// 将毫秒时间戳转换为可读格式
					addedAtTime := time.Unix(result.AddedAt/1000, 0)
					sb.WriteString(fmt.Sprintf("  ⏰ 添加时间: %s\n", addedAtTime.Format("2006-01-02 15:04:05")))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// SendUsersInfo 发送用户信息
func (bm *Manager) SendUsersInfo(chatID int64, messageID int) {
	allServers := bm.mediaServerManager.GetAllServers()
	var text string

	if len(allServers) == 0 {
		text = "没有找到媒体服务器"
	} else {
		text = "*👥 用户信息*:\n\n"
		for serverType, server := range allServers {
			users, err := server.GetUsers()
			if err != nil {
				text += fmt.Sprintf("*%s 服务器*:\n❌ 获取用户信息失败\n\n", strings.Title(string(serverType)))
				continue
			}

			text += fmt.Sprintf("*%s 服务器*:\n", strings.Title(string(serverType)))
			if len(users) == 0 {
				text += "暂无用户\n"
			} else {
				for _, user := range users {
					// 格式化最后在线时间
					lastSeen := "从未登录"
					if user.LastSeen > 0 {
						// lastSeen 是毫秒时间戳
						lastSeenTime := time.Unix(user.LastSeen/1000, 0).Format("2006-01-02 15:04:05")
						lastSeen = lastSeenTime
					}

					activeStatus := "❌ 非活跃"
					if user.IsActive {
						activeStatus = "✅ 活跃"
					}

					text += fmt.Sprintf("👤 *%s*\n", user.Username)
					text += fmt.Sprintf("   %s | %s\n", user.Type, activeStatus)
					text += fmt.Sprintf("   👀 最后在线: %s\n", lastSeen)
					text += "\n"
				}
			}
			text += "\n"
		}
	}

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = "Markdown"
		menu := CreateUsersInfoMenu()
		edit.ReplyMarkup = &menu
		err := editBotMessage(bm.Bot, edit)
		if err != nil {
			log.Printf("编辑用户信息消息失败: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = CreateUsersInfoMenu()
		err := sendBotMessage(bm.Bot, msg)
		if err != nil {
			log.Printf("发送用户信息消息失败: %v", err)
		}
	}
}

// SendMyStats 发送个人统计信息
func (bm *Manager) SendMyStats(chatID int64, messageID int) {
	allServers := bm.mediaServerManager.GetAllServers()
	var text string

	if len(allServers) == 0 {
		text = "📭 没有找到媒体服务器"
	} else {
		text = "*📈 个人统计信息*:\n\n"
		for serverType, server := range allServers {
			user, err := server.GetCurrentUser()
			if err != nil {
				text += fmt.Sprintf("*%s 服务器*:\n❌ 获取个人信息失败\n\n", strings.Title(string(serverType)))
				continue
			}

			stats, err := server.GetListeningStats()
			if err != nil {
				text += fmt.Sprintf("*%s 服务器*:\n❌ 获取统计信息失败\n\n", strings.Title(string(serverType)))
				continue
			}

			// 格式化最后在线时间
			lastSeen := "从未登录"
			if user.LastSeen > 0 {
				// lastSeen 是毫秒时间戳
				lastSeenTime := time.Unix(user.LastSeen/1000, 0).Format("2006-01-02 15:04:05")
				lastSeen = lastSeenTime
			}

			activeStatus := "❌ 非活跃"
			if user.IsActive {
				activeStatus = "✅ 活跃"
			}

			text += fmt.Sprintf("*%s 服务器*:\n", strings.Title(string(serverType)))
			text += fmt.Sprintf("👤 *%s*\n", user.Username)
			text += fmt.Sprintf("   %s | %s\n", user.Type, activeStatus)
			text += fmt.Sprintf("   👀 最后在线: %s\n", lastSeen)

			// 显示收听/观看统计
			if totalTime, ok := stats["TotalRecordCount"]; ok {
				text += fmt.Sprintf("   📊 统计: %v 个项目\n", totalTime)
			}
			text += "\n"
		}
	}

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = "Markdown"
		menu := CreateMyStatsMenu()
		edit.ReplyMarkup = &menu
		err := editBotMessage(bm.Bot, edit)
		if err != nil {
			log.Printf("编辑个人统计消息失败: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = CreateMyStatsMenu()
		err := sendBotMessage(bm.Bot, msg)
		if err != nil {
			log.Printf("发送个人统计消息失败: %v", err)
		}
	}
}

// EditHelpMessage 编辑帮助信息
func (bm *Manager) EditHelpMessage(chatID int64, messageID int) {
	helpText := `🎧 *多服务器媒体管理机器人帮助*

可用命令:
• /start - 显示主菜单
• /serverinfo - 获取所有服务器信息
• /users - 获取所有服务器的用户信息
• /libraries - 获取所有服务器的媒体库列表
• /search - 搜索所有服务器的媒体
• /mystats - 获取所有服务器的个人统计信息
• /help - 显示此帮助信息

或者使用下方的菜单按钮进行操作。
`
	edit := tgbotapi.NewEditMessageText(chatID, messageID, helpText)
	edit.ParseMode = "Markdown"
	menu := CreateMainMenu()
	edit.ReplyMarkup = &menu
	err := editBotMessage(bm.Bot, edit)
	if err != nil {
		log.Printf("编辑帮助信息消息失败: %v", err)
	}
}

// SendMessage 发送简单文本消息
func (bm *Manager) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	err := sendBotMessage(bm.Bot, msg)
	if err != nil {
		log.Printf("发送简单消息失败: %v", err)
	}
}

// EditMessage 编辑简单文本消息
func (bm *Manager) EditMessage(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	err := editBotMessage(bm.Bot, edit)
	if err != nil {
		log.Printf("编辑消息失败: %v", err)
	}
}

// sendBotMessage 发送消息并处理错误
func sendBotMessage(bot *tgbotapi.BotAPI, msg tgbotapi.Chattable) error {
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("发送消息失败: %v", err)
		return err
	}
	return nil
}

// editBotMessage 编辑消息并处理错误
func editBotMessage(bot *tgbotapi.BotAPI, edit tgbotapi.Chattable) error {
	_, err := bot.Send(edit)
	if err != nil {
		log.Printf("编辑消息失败: %v", err)
		return err
	}
	return nil
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

// executeWithLoadingStatus 在执行操作时显示加载状态
func executeWithLoadingStatus(bot *tgbotapi.BotAPI, chatID int64, messageID int, loadingText string, operation func()) {
	// 显示加载状态
	edit := tgbotapi.NewEditMessageText(chatID, messageID, loadingText)
	err := editBotMessage(bot, edit)
	if err != nil {
		log.Printf("发送加载状态失败: %v", err)
	}
	// 执行实际操作
	operation()
}

func initializeTelegramBot(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	var telegramBot *tgbotapi.BotAPI
	var err error

	// 如果设置了代理，则通过代理连接 Telegram
	if cfg.ProxyAddress != "" {
		log.Printf("使用代理连接 Telegram: %s", cfg.ProxyAddress)
		proxyURL, err := url.Parse("http://" + cfg.ProxyAddress)
		if err != nil {
			return nil, fmt.Errorf("无效的代理地址: %v", err)
		}

		proxyClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
			Timeout: 30 * time.Second,
		}

		telegramBot, err = tgbotapi.NewBotAPIWithClient(cfg.TelegramBotToken, tgbotapi.APIEndpoint, proxyClient)
		if err != nil {
			return nil, fmt.Errorf("无法通过代理连接到 Telegram Bot API: %v", err)
		}
	} else {
		telegramBot, err = tgbotapi.NewBotAPI(cfg.TelegramBotToken)
		if err != nil {
			return nil, fmt.Errorf("无法连接到 Telegram Bot API: %v", err)
		}
	}

	if cfg.Debug {
		telegramBot.Debug = true
	}

	return telegramBot, nil
}
