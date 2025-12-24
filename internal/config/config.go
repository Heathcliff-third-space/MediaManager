package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	TelegramBotToken    string
	AudiobookshelfURL   string
	AudiobookshelfToken string
	AudiobookshelfPort  int
	EmbyURL             string
	EmbyToken           string
	EmbyPort            int
	Debug               bool
	ProxyAddress        string
	AllowedUserIDs      []int64
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// 尝试在不同位置加载 .env 文件
	loadEnvFile()

	// 解析允许的用户ID列表
	allowedUserIDs := parseAllowedUserIDs(getEnvWithDefault("ALLOWED_USER_IDS", ""))

	config := &Config{
		TelegramBotToken:    getEnvWithDefault("TELEGRAM_BOT_TOKEN", ""),
		AudiobookshelfURL:   getEnvWithDefault("AUDIOBOOKSHELF_URL", ""),
		AudiobookshelfToken: getEnvWithDefault("AUDIOBOOKSHELF_TOKEN", ""),
		EmbyURL:             getEnvWithDefault("EMBY_URL", ""),
		EmbyToken:           getEnvWithDefault("EMBY_TOKEN", ""),
		Debug:               getEnvWithDefault("DEBUG", "false") == "true",
		ProxyAddress:        getEnvWithDefault("PROXY_ADDRESS", ""),
		AllowedUserIDs:      allowedUserIDs,
	}

	// 处理Audiobookshelf端口
	portStr := getEnvWithDefault("AUDIOBOOKSHELF_PORT", "")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			config.AudiobookshelfPort = port
		}
	}

	if config.AudiobookshelfPort == 0 {
		config.AudiobookshelfPort = 13378 // Audiobookshelf 默认端口
	}

	// 处理Emby端口
	embyPortStr := getEnvWithDefault("EMBY_PORT", "")
	if embyPortStr != "" {
		port, err := strconv.Atoi(embyPortStr)
		if err == nil {
			config.EmbyPort = port
		}
	}

	if config.EmbyPort == 0 {
		config.EmbyPort = 8096 // Emby 默认端口
	}

	return config
}

// loadEnvFile 尝试从不同位置加载 .env 文件
func loadEnvFile() {
	// 定义可能的 .env 文件路径，优先从 conf 目录加载
	possiblePaths := []string{
		"conf/.env",
		"./conf/.env",
		".env",
		"./.env",
		"../.env",
		"../../.env",
		"../../../.env",
	}

	// 获取当前工作目录
	currentDir, _ := os.Getwd()
	log.Printf("当前工作目录: %s", currentDir)

	// 尝试在当前目录及其上级目录中查找 .env 文件
	for _, path := range possiblePaths {
		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err == nil {
			log.Printf("尝试加载 .env 文件: %s", absPath)
			if err := godotenv.Load(absPath); err != nil {
				log.Printf("加载 .env 文件失败 (%s): %v", absPath, err)
			} else {
				log.Printf("成功加载 .env 文件: %s", absPath)
				return
			}
		} else {
			log.Printf(".env 文件不存在: %s", absPath)
		}
	}

	// 动态计算项目根目录
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	projectRoot := filepath.Join(basepath, "../../")
	projectEnv := filepath.Join(projectRoot, ".env")

	if _, err := os.Stat(projectEnv); err == nil {
		log.Printf("尝试加载 .env 文件: %s", projectEnv)
		if err := godotenv.Load(projectEnv); err != nil {
			log.Printf("加载 .env 文件失败 (%s): %v", projectEnv, err)
		} else {
			log.Printf("成功加载 .env 文件: %s", projectEnv)
			return
		}
	}

	log.Println("未找到 .env 文件，将使用系统环境变量")
}

// getEnvWithDefault 获取环境变量，如果不存在则返回默认值
func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseAllowedUserIDs 解析允许的用户ID列表
func parseAllowedUserIDs(idsStr string) []int64 {
	if idsStr == "" {
		return []int64{}
	}

	var ids []int64
	idStrings := strings.Split(idsStr, ",")
	for _, idStr := range idStrings {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}