package services

import (
	"fmt"
	"github.com/Heathcliff-third-space/MediaManager/internal/api"
	"github.com/Heathcliff-third-space/MediaManager/internal/config"
	"github.com/Heathcliff-third-space/MediaManager/internal/models"
	"sync"
)

// MediaServerType 媒体服务器类型
type MediaServerType string

const (
	AbsServerType MediaServerType = "audiobookshelf"
	EmbyServerType MediaServerType = "emby"
)

// MediaServerManager 管理多个媒体服务器
type MediaServerManager struct {
	servers map[MediaServerType]models.MediaServer
	mutex   sync.RWMutex
}

// NewMediaServerManager 创建新的媒体服务器管理器
func NewMediaServerManager(cfg *config.Config) (*MediaServerManager, error) {
	manager := &MediaServerManager{
		servers: make(map[MediaServerType]models.MediaServer),
	}

	// 初始化Audiobookshelf服务器
	if cfg.AudiobookshelfToken != "" {
		absClient := api.NewAbsClient(cfg)
		absAdapter := api.NewAbsAdapter(absClient)
		manager.servers[AbsServerType] = absAdapter
	}

	// 初始化Emby服务器
	if cfg.EmbyToken != "" {
		embyClient := api.NewEmbyClient(cfg)
		embyAdapter := api.NewEmbyAdapter(embyClient)
		manager.servers[EmbyServerType] = embyAdapter
	}

	if len(manager.servers) == 0 {
		return nil, fmt.Errorf("没有配置任何媒体服务器")
	}

	return manager, nil
}

// GetServer 获取指定类型的媒体服务器
func (m *MediaServerManager) GetServer(serverType MediaServerType) (models.MediaServer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	server, exists := m.servers[serverType]
	if !exists {
		return nil, fmt.Errorf("未找到类型为 %s 的媒体服务器", serverType)
	}

	return server, nil
}

// GetAllServers 获取所有可用的媒体服务器
func (m *MediaServerManager) GetAllServers() map[MediaServerType]models.MediaServer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	servers := make(map[MediaServerType]models.MediaServer)
	for k, v := range m.servers {
		servers[k] = v
	}

	return servers
}

// GetServerTypes 获取所有可用的服务器类型
func (m *MediaServerManager) GetServerTypes() []MediaServerType {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var types []MediaServerType
	for serverType := range m.servers {
		types = append(types, serverType)
	}

	return types
}

// SearchAcrossServers 在所有服务器中搜索
func (m *MediaServerManager) SearchAcrossServers(query string) (map[MediaServerType][]models.SearchResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	results := make(map[MediaServerType][]models.SearchResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 使用信号量控制最大并发数
	maxConcurrency := make(chan struct{}, 4)

	for serverType, server := range m.servers {
		wg.Add(1)
		go func(st MediaServerType, s models.MediaServer) {
			defer wg.Done()
			// 控制并发数
			maxConcurrency <- struct{}{}
			defer func() { <-maxConcurrency }()

			searchResults, err := s.Search(query)
			if err != nil {
				// 记录错误但继续处理其他服务器
				return
			}

			mu.Lock()
			results[st] = searchResults
			mu.Unlock()
		}(serverType, server)
	}

	wg.Wait()

	return results, nil
}

// GetServerInfoAcrossServers 获取所有服务器的信息
func (m *MediaServerManager) GetServerInfoAcrossServers() (map[MediaServerType]*models.ServerInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	info := make(map[MediaServerType]*models.ServerInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 使用信号量控制最大并发数
	maxConcurrency := make(chan struct{}, 4)

	for serverType, server := range m.servers {
		wg.Add(1)
		go func(st MediaServerType, s models.MediaServer) {
			defer wg.Done()
			// 控制并发数
			maxConcurrency <- struct{}{}
			defer func() { <-maxConcurrency }()

			serverInfo, err := s.GetServerInfo()
			if err != nil {
				// 记录错误但继续处理其他服务器
				return
			}

			mu.Lock()
			info[st] = serverInfo
			mu.Unlock()
		}(serverType, server)
	}

	wg.Wait()

	return info, nil
}