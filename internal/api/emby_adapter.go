package api

import (
	"encoding/json"
	"fmt"
	"github.com/Heathcliff-third-space/MediaManager/internal/models"
	"strings"
	"sync"
	"time"
)

// EmbyAdapter 实现 MediaServer 接口，作为 Emby 的适配器
type EmbyAdapter struct {
	client *EmbyClient
	// 添加媒体库缓存
	librariesCache      []models.LibraryInfo
	librariesCacheTime  time.Time
	librariesCacheMutex sync.RWMutex
	cacheExpiry         time.Duration
}

// NewEmbyAdapter 创建新的 Emby 适配器
func NewEmbyAdapter(client *EmbyClient) *EmbyAdapter {
	return &EmbyAdapter{
		client: client,
	}
}

// GetServerInfo 实现 MediaServer 接口
func (e *EmbyAdapter) GetServerInfo() (*models.ServerInfo, error) {
	data, err := e.client.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	var systemInfo struct {
		ID                  string `json:"Id"`
		ServerName          string `json:"ServerName"`
		Version             string `json:"Version"`
		OperatingSystem     string `json:"OperatingSystem"`
		Architecture        string `json:"Architecture"`
		HasHttps            bool   `json:"HasHttps"`
		UserId              string `json:"UserId"`
		LocalAddress        string `json:"LocalAddress"`
		WanAddress          string `json:"WanAddress"`
		IsShuttingDown      bool   `json:"IsShuttingDown"`
		SupportsHttps       bool   `json:"SupportsHttps"`
		HttpsPortNumber     int    `json:"HttpsPortNumber"`
		TranscodingId       string `json:"TranscodingId"`
		WebSocketPortNumber int    `json:"WebSocketPortNumber"`
		CanSelfRestart      bool   `json:"CanSelfRestart"`
	}

	err = json.Unmarshal(data, &systemInfo)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling system info: %w", err)
	}

	serverInfo := &models.ServerInfo{
		ID:            systemInfo.ID,
		Name:          systemInfo.ServerName,
		Version:       systemInfo.Version,
		OS:            systemInfo.OperatingSystem,
		Arch:          systemInfo.Architecture,
		ServerVersion: systemInfo.Version,
		// Emby没有直接的API版本字段
		APIVersion: "Emby",
	}

	return serverInfo, nil
}

// GetUsers 实现 MediaServer 接口
func (e *EmbyAdapter) GetUsers() ([]models.UserInfo, error) {
	data, err := e.client.GetUsers()
	if err != nil {
		return nil, err
	}

	var embyUsers []models.EmbyUser
	err = json.Unmarshal(data, &embyUsers)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling users: %w", err)
	}

	// 转换Emby用户信息到通用用户信息
	users := make([]models.UserInfo, len(embyUsers))
	for i, embyUser := range embyUsers {
		users[i] = *toUser(&embyUser)
	}

	return users, nil
}

// GetCurrentUser 实现 MediaServer 接口
func (e *EmbyAdapter) GetCurrentUser() (*models.UserInfo, error) {
	data, err := e.client.GetCurrentUser()
	if err != nil {
		return nil, err
	}

	var embyUser models.EmbyUser

	err = json.Unmarshal(data, &embyUser)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling current user: %w", err)
	}

	user := toUser(&embyUser)

	return user, nil
}

func toUser(embyUser *models.EmbyUser) *models.UserInfo {
	var lastSeen int64 = 0
	if embyUser.LastActivityDate != "" {
		// 解析ISO 8601时间格式
		t, err := time.Parse(time.RFC3339, embyUser.LastActivityDate)
		if err == nil {
			lastSeen = t.Unix() * 1000 // 转换为毫秒时间戳
		}
	}

	userType := "EmbyUser"
	if embyUser.Policy.IsAdministrator {
		userType = "Admin"
	}

	user := &models.UserInfo{
		ID:       embyUser.ID,
		Username: embyUser.Name,
		Type:     userType,
		IsActive: !embyUser.Policy.IsDisabled,
		LastSeen: lastSeen,
		// Emby用户信息中没有明确的创建时间，使用0值
		CreatedAt: 0,
		UpdatedAt: 0,
	}
	return user
}

// GetLibraries 实现 MediaServer 接口
func (e *EmbyAdapter) GetLibraries() ([]models.LibraryInfo, error) {
	data, err := e.client.GetMediaFolders()
	if err != nil {
		return nil, err
	}

	var mediaFolders struct {
		Items []struct {
			Name            string `json:"Name"`
			ID              string `json:"Id"`
			CollectionType  string `json:"CollectionType"`
			LocationType    string `json:"LocationType"`
			RefreshStatus   string `json:"RefreshStatus"`
			PrimaryImageTag string `json:"PrimaryImageTag"`
			CanDelete       bool   `json:"CanDelete"`
			CanRefresh      bool   `json:"CanRefresh"`
			CanRecover      bool   `json:"CanRecover"`
			CanDownload     bool   `json:"CanDownload"`
			IsOwnedItem     bool   `json:"IsOwnedItem"`
		} `json:"Items"`
	}

	err = json.Unmarshal(data, &mediaFolders)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling media folders: %w", err)
	}

	// 转换Emby媒体库信息到通用媒体库信息
	libraries := make([]models.LibraryInfo, len(mediaFolders.Items))
	for i, folder := range mediaFolders.Items {
		// 获取媒体库项目数量
		itemCount, err := e.client.GetLibraryItemsCount(folder.ID)
		if err != nil {
			// 如果获取项目数量失败，设置为0
			itemCount = 0
		}
		libraries[i] = models.LibraryInfo{
			ID:        folder.ID,
			Name:      folder.Name,
			ItemCount: itemCount,
			MediaType: folder.CollectionType,
			// Emby媒体库信息中没有明确的创建时间，使用0值
			CreatedAt: 0,
			UpdatedAt: 0,
			LastScan:  0, // Emby中可以获取最后扫描时间，这里简化处理
		}
	}

	return libraries, nil
}

// GetLibraryItemsCount 实现 MediaServer 接口
func (e *EmbyAdapter) GetLibraryItemsCount(libraryID string) (int, error) {
	data, err := e.client.GetItems(libraryID, "", nil)
	if err != nil {
		return 0, err
	}

	var itemsResponse struct {
		Items            []interface{} `json:"Items"`
		TotalRecordCount int           `json:"TotalRecordCount"`
	}

	err = json.Unmarshal(data, &itemsResponse)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling items: %w", err)
	}

	return itemsResponse.TotalRecordCount, nil
}

// Search 实现 MediaServer 接口
func (e *EmbyAdapter) Search(query string) ([]models.SearchResult, error) {
	data, err := e.client.SearchItems(query, "", 50) // 限制返回50个结果
	if err != nil {
		return nil, err
	}

	var searchResponse struct {
		Items []struct {
			ID             string   `json:"Id"`
			Name           string   `json:"Name"`
			Type           string   `json:"Type"`
			IsFolder       bool     `json:"IsFolder"`
			Size           int64    `json:"Size"`
			DateCreated    string   `json:"DateCreated"`
			ParentId       string   `json:"ParentId"`
			Path           string   `json:"Path"`
			ProductionYear int      `json:"ProductionYear"`
			PremiereDate   string   `json:"PremiereDate"`
			Overview       string   `json:"Overview"`
			Genres         []string `json:"Genres"`
			MediaType      string   `json:"MediaType"`
			RunTimeTicks   int64    `json:"RunTimeTicks"`
			ProviderIds    struct {
				Tmdb string `json:"Tmdb"`
				Imdb string `json:"Imdb"`
				Tvdb string `json:"Tvdb"`
			} `json:"ProviderIds"`
		} `json:"Items"`
	}

	err = json.Unmarshal(data, &searchResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling search results: %w", err)
	}

	// 转换Emby搜索结果到通用搜索结果
	results := make([]models.SearchResult, len(searchResponse.Items))
	for i, item := range searchResponse.Items {
		var addedAt int64 = 0
		// 解析ISO 8601时间格式
		if item.DateCreated != "" {
			t, err := time.Parse(time.RFC3339, item.DateCreated)
			if err == nil {
				addedAt = t.Unix() * 1000 // 转换为毫秒时间戳
			}
		}

		// 从路径中提取库信息 - 实际中需要获取库名称
		library := item.ParentId // 实际中需要获取库名称
		if library == "" {
			library = "Unknown Library"
		}

		results[i] = models.SearchResult{
			ID:             item.ID,
			Title:          item.Name,
			Author:         "", // Emby通常没有作者字段，除非是audiobook类型
			Size:           item.Size,
			AddedAt:        addedAt,
			LibraryID:      item.ParentId,
			Library:        library,
			Type:           strings.ToLower(item.Type),
			Path:           item.Path,
			RelPath:        item.Path,
			Overview:       item.Overview,
			Genres:         item.Genres,
			Year:           item.ProductionYear,
			ProductionYear: item.ProductionYear,
			PremiereDate:   item.PremiereDate,
			RunTime:        item.RunTimeTicks,
			MediaType:      item.MediaType,
		}
	}

	return results, nil
}

// GetListeningStats 实现 MediaServer 接口
func (e *EmbyAdapter) GetListeningStats() (map[string]interface{}, error) {
	// 获取当前用户信息以获取用户ID
	currentUser, err := e.GetCurrentUser()
	if err != nil {
		return nil, err
	}

	// 获取用户的媒体播放进度
	data, err := e.client.GetUserData(currentUser.ID)
	if err != nil {
		return nil, err
	}

	var stats map[string]interface{}
	err = json.Unmarshal(data, &stats)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling listening stats: %w", err)
	}

	return stats, nil
}
