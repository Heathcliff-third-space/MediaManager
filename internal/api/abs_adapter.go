package api

import (
	"fmt"
	"github.com/Heathcliff-third-space/MediaManager/internal/models"
	"sync"
	"time"
)

// AbsAdapter 实现 MediaServer 接口，作为 Audiobookshelf 的适配器
type AbsAdapter struct {
	client *AbsClient
	// 添加媒体库缓存
	librariesCache      []models.LibraryInfo
	librariesCacheTime  time.Time
	librariesCacheMutex sync.RWMutex
	cacheExpiry         time.Duration
}

// NewAbsAdapter 创建新的 Audiobookshelf 适配器
func NewAbsAdapter(client *AbsClient) *AbsAdapter {
	return &AbsAdapter{
		client:      client,
		cacheExpiry: 30 * time.Minute, // 默认30分钟缓存过期时间
	}
}

// GetServerInfo 实现 MediaServer 接口
func (a *AbsAdapter) GetServerInfo() (*models.ServerInfo, error) {
	status, err := a.client.GetServerStatus()
	if err != nil {
		return nil, err
	}

	// 由于Audiobookshelf的status API不提供完整的服务器信息，
	// 我们只设置基本的服务器信息
	serverInfo := &models.ServerInfo{
		Version:       status.ServerVersion,
		APIVersion:    status.APIVersion,
		Language:      status.Language,
		ServerVersion: status.ServerVersion,
	}

	return serverInfo, nil
}

// GetUsers 实现 MediaServer 接口
func (a *AbsAdapter) GetUsers() ([]models.UserInfo, error) {
	absUsers, err := a.client.GetUsers()
	if err != nil {
		return nil, err
	}

	// 转换Audiobookshelf用户信息到通用用户信息
	users := make([]models.UserInfo, len(absUsers))

	for i, absUser := range absUsers {
		users[i] = models.UserInfo{
			ID:        absUser.ID,
			Username:  absUser.Username,
			Type:      absUser.Type,
			IsActive:  absUser.IsActive,
			LastSeen:  absUser.LastSeen,
			CreatedAt: absUser.CreatedAt,
			UpdatedAt: absUser.UpdatedAt,
		}
	}

	return users, nil
}

// GetCurrentUser 实现 MediaServer 接口
func (a *AbsAdapter) GetCurrentUser() (*models.UserInfo, error) {
	absUser, err := a.client.GetCurrentUser()
	if err != nil {
		return nil, err
	}

	// 转换Audiobookshelf用户信息到通用用户信息
	user := &models.UserInfo{
		ID:        absUser.ID,
		Username:  absUser.Username,
		Type:      absUser.Type,
		IsActive:  absUser.IsActive,
		LastSeen:  absUser.LastSeen,
		CreatedAt: absUser.CreatedAt,
		UpdatedAt: absUser.UpdatedAt,
	}

	return user, nil
}

// GetLibraries 实现 MediaServer 接口
func (a *AbsAdapter) GetLibraries() ([]models.LibraryInfo, error) {
	absLibraries, err := a.client.GetLibrariesInfo()
	if err != nil {
		return nil, err
	}

	// 转换Audiobookshelf媒体库信息到通用媒体库信息
	libraries := make([]models.LibraryInfo, len(absLibraries))
	for i, absLibrary := range absLibraries {
		libraries[i] = models.LibraryInfo{
			ID:        absLibrary.ID,
			Name:      absLibrary.Name,
			ItemCount: absLibrary.ItemCount,
			MediaType: absLibrary.MediaType,
			CreatedAt: absLibrary.CreatedAt,
			UpdatedAt: absLibrary.UpdatedAt,
			LastScan:  absLibrary.LastScan,
		}
	}

	return libraries, nil
}

// GetLibraryItemsCount 实现 MediaServer 接口
func (a *AbsAdapter) GetLibraryItemsCount(libraryID string) (int, error) {
	return a.client.GetLibraryItemsCount(libraryID)
}

// Search 实现 MediaServer 接口
func (a *AbsAdapter) Search(query string) ([]models.SearchResult, error) {
	books, err := a.client.SearchBooks(query, "")
	if err != nil {
		return nil, err
	}

	// 转换Audiobookshelf搜索结果到通用搜索结果
	results := make([]models.SearchResult, len(books))
	for i, book := range books {
		// 从相对路径中提取标题（通常是文件名）
		title := book.RelPath
		if len(title) > 0 {
			// 提取文件名部分作为标题
			title = extractFileName(book.RelPath)
		}

		// 获取媒体库名称，如果无法获取则使用ID
		libraryName := "Unknown Library"
		if book.LibraryID != "" {
			// 尝试获取媒体库名称
			libraryName = a.getLibraryNameByID(book.LibraryID)
			if libraryName == "" {
				libraryName = book.LibraryID // 如果无法获取名称，使用ID
			}
		}

		results[i] = models.SearchResult{
			ID:      fmt.Sprintf("%s_%s", book.LibraryID, book.RelPath),
			Title:   title,
			Size:    book.Size,
			AddedAt: book.AddedAt,
			// 尝试获取媒体库名称，如果获取失败则使用ID
			LibraryID:      book.LibraryID,
			Library:        libraryName,
			Type:           "book", // Audiobookshelf主要是书籍
			RelPath:        book.RelPath,
			Overview:       "",         // Audiobookshelf搜索结果中没有概述信息
			Genres:         []string{}, // Audiobookshelf搜索结果中没有分类信息
			Year:           0,          // Audiobookshelf搜索结果中没有年份信息
			ProductionYear: 0,          // Audiobookshelf搜索结果中没有制作年份信息
			PremiereDate:   "",         // Audiobookshelf搜索结果中没有首映日期
			RunTime:        0,          // Audiobookshelf搜索结果中没有运行时间
			MediaType:      "audio",    // Audiobookshelf主要是音频媒体
		}
	}

	return results, nil
}

// GetListeningStats 实现 MediaServer 接口
func (a *AbsAdapter) GetListeningStats() (map[string]interface{}, error) {
	return a.client.GetListeningStats()
}

// getLibraryNameByID 根据ID获取媒体库名称
func (a *AbsAdapter) getLibraryNameByID(libraryID string) string {
	// 检查缓存
	a.librariesCacheMutex.RLock()
	if time.Since(a.librariesCacheTime) < a.cacheExpiry && a.librariesCache != nil {
		// 在缓存中查找
		for _, lib := range a.librariesCache {
			if lib.ID == libraryID {
				a.librariesCacheMutex.RUnlock()
				return lib.Name
			}
		}
		a.librariesCacheMutex.RUnlock()
		return "" // 缓存中有数据但没找到对应ID
	}
	a.librariesCacheMutex.RUnlock()

	// 缓存未命中或已过期，获取新的媒体库列表
	a.librariesCacheMutex.Lock()
	// 双重检查，防止并发问题
	if time.Since(a.librariesCacheTime) >= a.cacheExpiry || a.librariesCache == nil {
		libraries, err := a.client.GetLibrariesInfo()
		if err == nil {
			// 转换Audiobookshelf媒体库信息到通用媒体库信息
			a.librariesCache = make([]models.LibraryInfo, len(libraries))
			for i, absLibrary := range libraries {
				a.librariesCache[i] = models.LibraryInfo{
					ID:        absLibrary.ID,
					Name:      absLibrary.Name,
					ItemCount: absLibrary.ItemCount,
					MediaType: absLibrary.MediaType,
					CreatedAt: absLibrary.CreatedAt,
					UpdatedAt: absLibrary.UpdatedAt,
					LastScan:  absLibrary.LastScan,
				}
			}
			a.librariesCacheTime = time.Now()
		} else {
			// 获取失败，更新缓存时间以避免频繁重试
			a.librariesCacheTime = time.Now()
		}
	}
	// 在新获取的数据中查找
	for _, lib := range a.librariesCache {
		if lib.ID == libraryID {
			a.librariesCacheMutex.Unlock()
			return lib.Name
		}
	}
	a.librariesCacheMutex.Unlock()
	return "" // 没有找到
}

// 辅助函数：从路径中提取文件名
func extractFileName(path string) string {
	// 简单实现：取路径中最后一个/后面的部分
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
