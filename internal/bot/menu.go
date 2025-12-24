package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RegisterCommands 注册 Telegram Bot 命令
func RegisterCommands(bot *tgbotapi.BotAPI) error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "显示主菜单"},
		{Command: "serverinfo", Description: "获取所有服务器信息"},
		{Command: "users", Description: "获取所有服务器的用户列表"},
		{Command: "libraries", Description: "获取所有服务器的媒体库列表"},
		{Command: "search", Description: "搜索所有服务器的媒体"},
		{Command: "mystats", Description: "获取所有服务器的个人统计信息"},
		{Command: "help", Description: "显示帮助信息"},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := bot.Request(config)
	return err
}

// CreateMainMenu 创建主菜单
func CreateMainMenu() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("📊 服务器信息", "system_info"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("👥 用户列表", "users_list"),
			tgbotapi.NewInlineKeyboardButtonData("📚 媒体库", "libraries_list"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🔍 搜索媒体", "search_books"),
			tgbotapi.NewInlineKeyboardButtonData("📈 我的统计", "my_stats"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "help"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateServerInfoMenu 创建服务器信息菜单
func CreateServerInfoMenu() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("⬅ 返回主菜单", "main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateUsersInfoMenu 创建用户信息菜单
func CreateUsersInfoMenu() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("⬅ 返回主菜单", "main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateLibrariesMenu 创建媒体库菜单
func CreateLibrariesMenu() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("⬅ 返回主菜单", "main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateSearchMenu 创建搜索菜单
func CreateSearchMenu() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("⬅ 返回主菜单", "main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// CreateMyStatsMenu 创建我的统计菜单
func CreateMyStatsMenu() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("⬅ 返回主菜单", "main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}