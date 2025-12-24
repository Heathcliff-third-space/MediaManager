package services

import (
	"fmt"
	"github.com/Heathcliff-third-space/MediaManager/internal/api"
	"github.com/Heathcliff-third-space/MediaManager/internal/models"
	"github.com/Heathcliff-third-space/MediaManager/internal/util"
	"strings"
	"sync"
	"time"
)

// AbsServerService 服务器信息服务
type AbsServerService struct {
	adapter *api.AbsAdapter
	// 添加缓存相关字段
	librariesCache      []LibraryWithStats
	librariesCacheTime  time.Time
	librariesCacheMutex sync.RWMutex
	cacheExpiry         time.Duration
}

// NewAbsServerService 创建服务器信息服务实例
func NewAbsServerService(adapter *api.AbsAdapter) *AbsServerService {
	return &AbsServerService{
		adapter:     adapter,
		cacheExpiry: 5 * time.Minute, // 默认5分钟缓存过期时间
	}
}

// GetFormattedServerInfo 获取格式化的服务器信息
func (s *AbsServerService) GetFormattedServerInfo() (string, error) {
	status, err := s.adapter.GetServerInfo()
	if err != nil {
		return "", fmt.Errorf("获取服务器状态失败: %w", err)
	}

	// 格式化服务器信息
	var sb strings.Builder

	sb.WriteString("📊 *Audiobookshelf 服务器信息*\n\n")

	// 注意：AbsServerStatus 模型中没有 App 字段，使用 ServerVersion 替代
	sb.WriteString(fmt.Sprintf("🖥 *版本*: `%s`\n", status.ServerVersion))
	sb.WriteString(fmt.Sprintf("🔤 *语言*: `%s`\n", status.Language))

	sb.WriteString("\n📚 *媒体库信息*\n")

	// 获取媒体库信息
	libraries, err := s.GetLibrariesWithStats()
	if err != nil {
		sb.WriteString("⚠️ 获取媒体库信息失败\n")
	} else {
		if len(libraries) == 0 {
			sb.WriteString("⚠️ 暂无媒体库\n")
		} else {
			sb.WriteString(fmt.Sprintf("📁 媒体库总数: `%d`\n", len(libraries)))
			for _, lib := range libraries {
				sb.WriteString(fmt.Sprintf("%s %s (📚 %d)\n", util.GetMediaTypeIcon(lib.MediaType), lib.Name, lib.ItemCount))
			}
		}
	}

	return sb.String(), nil
}

// LibraryWithStats 带有统计信息的媒体库
type LibraryWithStats struct {
	models.LibraryInfo
	ItemCount int `json:"item_count"`
}

// GetLibrariesWithStats 获取带有统计信息的媒体库列表，带缓存功能
func (s *AbsServerService) GetLibrariesWithStats() ([]LibraryWithStats, error) {
	// 检查缓存
	s.librariesCacheMutex.RLock()
	if time.Since(s.librariesCacheTime) < s.cacheExpiry && s.librariesCache != nil {
		cached := s.librariesCache
		s.librariesCacheMutex.RUnlock()
		return cached, nil
	}
	s.librariesCacheMutex.RUnlock()

	// 缓存失效，获取新数据
	libraries, err := s.adapter.GetLibraries()
	if err != nil {
		return nil, err
	}

	// 获取每个库的详细统计信息，使用并行处理提高性能
	librariesWithStats := make([]LibraryWithStats, len(libraries))

	// 使用并行处理，最大并发数为4
	const maxConcurrency = 4
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并行获取每个库的统计信息
	for i, library := range libraries {
		wg.Add(1)
		go func(index int, lib models.LibraryInfo) {
			defer wg.Done()

			// 控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 获取库中媒体项的数量
			count, err := s.adapter.GetLibraryItemsCount(lib.ID)
			if err != nil {
				// 如果获取失败，设置为0
				mu.Lock()
				librariesWithStats[index].LibraryInfo = lib
				librariesWithStats[index].ItemCount = 0
				mu.Unlock()
			} else {
				mu.Lock()
				librariesWithStats[index].LibraryInfo = lib
				librariesWithStats[index].ItemCount = count
				mu.Unlock()
			}
		}(i, library)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 更新缓存
	s.librariesCacheMutex.Lock()
	s.librariesCache = librariesWithStats
	s.librariesCacheTime = time.Now()
	s.librariesCacheMutex.Unlock()

	return librariesWithStats, nil
}

// GetLibraryName 根据libraryId获取媒体库名称
func (s *AbsServerService) GetLibraryName(libraryId string) (string, error) {
	// 使用轻量级方法获取媒体库名称，避免获取统计信息
	libraries, err := s.getLibrariesBasicInfo()
	if err != nil {
		return "", err
	}

	// 查找指定ID的媒体库
	for _, lib := range libraries {
		if lib.ID == libraryId {
			return lib.Name, nil
		}
	}

	// 如果没有找到对应的媒体库，返回空字符串
	return "", fmt.Errorf("未找到ID为%s的媒体库", libraryId)
}

// getLibrariesBasicInfo 获取媒体库基本信息（ID和名称），不包含统计信息
func (s *AbsServerService) getLibrariesBasicInfo() ([]models.LibraryInfo, error) {
	// 直接调用API获取媒体库信息，不计算统计信息
	return s.adapter.GetLibraries()
}

// GetUsersWithProgress 获取用户列表及播放统计信息
func (s *AbsServerService) GetUsersWithProgress() ([]models.UserInfo, error) {
	// 获取用户列表
	users, err := s.adapter.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 获取每个用户的播放统计信息
	// 使用并行处理，最大并发数为4
	const maxConcurrency = 4
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range users {
		wg.Add(1)
		go func(index int, user models.UserInfo) {
			defer wg.Done()

			// 控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 获取用户的播放进度信息
			// 由于适配器接口没有直接提供用户进度，我们在这里可能需要特别处理
			// 目前简化处理，后续根据实际需要调整
			mu.Lock()
			users[index] = user
			mu.Unlock()
		}(i, users[i])
	}

	// 等待所有goroutine完成
	wg.Wait()

	return users, nil
}

// SearchBooks 搜索图书，使用并行处理提高性能
func (s *AbsServerService) SearchBooks(term string, libraryID string) ([]models.AbsBook, error) {
	if term == "" {
		return nil, fmt.Errorf("搜索词不能为空")
	}

	// 如果需要在特定媒体库中搜索，需要创建一个新方法
	// 目前适配器接口只支持全库搜索，我们暂时使用全库搜索结果并过滤
	results, err := s.adapter.Search(term)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 转换回AbsBook格式以保持兼容性
	books := make([]models.AbsBook, len(results))
	for i, result := range results {
		books[i] = models.AbsBook{
			LibraryID: result.LibraryID,
			RelPath:   result.RelPath,
			Size:      result.Size,
			AddedAt:   result.AddedAt,
		}
	}

	return books, nil
}

// GetCurrentUserWithProgress 获取当前用户信息及播放统计
func (s *AbsServerService) GetCurrentUserWithProgress() (*models.UserInfo, error) {
	// 获取当前用户信息
	user, err := s.adapter.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("获取当前用户信息失败: %w", err)
	}

	// 获取用户的播放进度信息
	// 由于适配器接口没有直接提供用户进度，我们在这里可能需要特别处理
	// 目前简化处理，后续根据实际需要调整
	return user, nil
}

// GetListeningStats 获取当前用户的收听统计信息
func (s *AbsServerService) GetListeningStats() (map[string]interface{}, error) {
	stats, err := s.adapter.GetListeningStats()
	if err != nil {
		return nil, fmt.Errorf("获取收听统计信息失败: %w", err)
	}
	return stats, nil
}
