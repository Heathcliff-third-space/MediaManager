package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Heathcliff-third-space/MediaManager/internal/config"
	"github.com/Heathcliff-third-space/MediaManager/internal/models"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// AbsClient represents an Audiobookshelf API client
type AbsClient struct {
	baseURL    string
	token      string
	httpClient *http.Client

	// 添加缓存相关字段
	librariesCache      []models.AbsLibraryInfo
	librariesCacheTime  time.Time
	librariesCacheMutex sync.RWMutex
	cacheExpiry         time.Duration
}

// NewAbsClient creates a new Audiobookshelf API client
func NewAbsClient(config *config.Config) *AbsClient {
	baseURL := config.AudiobookshelfURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", config.AudiobookshelfPort)
	}

	// 创建不使用代理的 HTTP 客户端
	client := &http.Client{}

	return &AbsClient{
		baseURL:     baseURL,
		token:       config.AudiobookshelfToken,
		httpClient:  client,
		cacheExpiry: 30 * time.Minute, // 默认30分钟缓存过期时间
	}
}

// DoRequestRaw performs an HTTP request to the Audiobookshelf API and returns raw response
func (c *AbsClient) DoRequestRaw(method, path string, body interface{}) ([]byte, error) {
	return c.doRequest(method, path, body)
}

// doRequest performs an HTTP request to the Audiobookshelf API
func (c *AbsClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetLibraries retrieves the list of libraries from Audiobookshelf
func (c *AbsClient) GetLibraries() ([]byte, error) {
	return c.doRequest("GET", "/api/libraries", nil)
}

// GetServerStatus 获取服务器状态信息
func (c *AbsClient) GetServerStatus() (*models.AbsServerStatus, error) {
	data, err := c.doRequest("GET", "/status", nil)
	if err != nil {
		return nil, err
	}

	var status models.AbsServerStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling server status: %w", err)
	}

	return &status, nil
}

// GetLibraryItemsCount 获取指定库中的媒体项数量
func (c *AbsClient) GetLibraryItemsCount(libraryID string) (int, error) {
	endpoint := fmt.Sprintf("/api/libraries/%s/items", libraryID)
	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}

	var response struct {
		Total int `json:"total"`
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling library items: %w", err)
	}

	return response.Total, nil
}

// GetLibrariesInfo 获取媒体库详细信息
func (c *AbsClient) GetLibrariesInfo() ([]models.AbsLibraryInfo, error) {
	// 检查缓存
	c.librariesCacheMutex.RLock()
	if time.Since(c.librariesCacheTime) < c.cacheExpiry && c.librariesCache != nil {
		cached := c.librariesCache
		c.librariesCacheMutex.RUnlock()
		return cached, nil
	}
	c.librariesCacheMutex.RUnlock()

	// 缓存失效，从API获取新数据
	data, err := c.doRequest("GET", "/api/libraries", nil)
	if err != nil {
		return nil, err
	}

	// API返回的是一个包含libraries字段的对象
	var response struct {
		Libraries []models.AbsLibraryInfo `json:"libraries"`
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling libraries: %w", err)
	}

	// 获取每个媒体库的项目数量
	for i := range response.Libraries {
		count, err := c.GetLibraryItemsCount(response.Libraries[i].ID)
		if err == nil {
			response.Libraries[i].ItemCount = count
		}
	}

	// 更新缓存
	c.librariesCacheMutex.Lock()
	c.librariesCache = response.Libraries
	c.librariesCacheTime = time.Now()
	c.librariesCacheMutex.Unlock()

	return response.Libraries, nil
}

// SearchBooks 搜索图书，支持并行处理
func (c *AbsClient) SearchBooks(term string, libraryID string) ([]models.AbsBook, error) {
	params := url.Values{}
	params.Add("q", term)

	// 如果指定了特定的媒体库ID，则只搜索该库
	if libraryID != "" {
		endpoint := fmt.Sprintf("/api/libraries/%s/search?%s", libraryID, params.Encode())

		data, err := c.doRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Results []struct {
				LibraryItem struct {
					Path    string `json:"path"`
					RelPath string `json:"relPath"`
					Size    int64  `json:"size"`
					AddedAt int64  `json:"addedAt"`
				}
			} `json:"book"`
		}

		err = json.Unmarshal(data, &response)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling search results: %w", err)
		}

		// 提取libraryItem中的字段
		var books []models.AbsBook
		for _, result := range response.Results {
			books = append(books, models.AbsBook{
				LibraryID: libraryID,
				RelPath:   result.LibraryItem.RelPath,
				Size:      result.LibraryItem.Size,
				AddedAt:   result.LibraryItem.AddedAt,
			})
		}

		return books, nil
	}

	// 如果没有指定特定的媒体库ID，则搜索所有库，使用并行处理
	libraries, err := c.GetLibrariesInfo()
	if err != nil {
		return nil, fmt.Errorf("获取媒体库列表失败: %w", err)
	}

	// 使用并行处理，最大并发数为4
	const maxConcurrency = 4
	semaphore := make(chan struct{}, maxConcurrency)

	var allBooks []models.AbsBook
	var mu sync.Mutex
	var wg sync.WaitGroup
	// 使用relPath作为唯一标识符进行去重
	bookRelPaths := make(map[string]bool)

	for _, lib := range libraries {
		wg.Add(1)
		go func(lib models.AbsLibraryInfo) {
			defer wg.Done()

			// 控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			endpoint := fmt.Sprintf("/api/libraries/%s/search?%s", lib.ID, params.Encode())

			data, err := c.doRequest("GET", endpoint, nil)
			if err != nil {
				// 继续搜索下一个库而不是完全失败
				return
			}

			var response struct {
				Results []struct {
					LibraryItem struct {
						Path    string `json:"path"`
						RelPath string `json:"relPath"`
						Size    int64  `json:"size"`
						AddedAt int64  `json:"addedAt"`
					}
				} `json:"book"`
			}

			err = json.Unmarshal(data, &response)
			if err != nil {
				// 继续搜索下一个库而不是完全失败
				return
			}

			// 添加去重逻辑
			mu.Lock()
			defer mu.Unlock()
			for _, result := range response.Results {
				if !bookRelPaths[result.LibraryItem.RelPath] {
					allBooks = append(allBooks, models.AbsBook{
						LibraryID: lib.ID,
						RelPath:   result.LibraryItem.RelPath,
						Size:      result.LibraryItem.Size,
						AddedAt:   result.LibraryItem.AddedAt,
					})
					bookRelPaths[result.LibraryItem.RelPath] = true
				}
			}
		}(lib)
	}

	// 等待所有goroutine完成
	wg.Wait()

	return allBooks, nil
}

// GetUsers 获取用户列表
func (c *AbsClient) GetUsers() ([]models.AbsUserInfo, error) {
	data, err := c.doRequest("GET", "/api/users", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Users []models.AbsUserInfo `json:"users"`
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling users: %w", err)
	}

	return response.Users, nil
}

// GetUserMediaProgress 获取用户的媒体播放进度信息
func (c *AbsClient) GetUserMediaProgress(userID string) ([]interface{}, error) {
	// 使用 /api/me/listening-stats 端点获取当前用户的收听统计
	data, err := c.doRequest("GET", "/api/me/listening-stats", nil)
	if err != nil {
		return nil, err
	}

	// 解析收听统计数据
	var stats map[string]interface{}
	err = json.Unmarshal(data, &stats)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling listening stats: %w", err)
	}

	// 从收听统计中提取媒体进度信息
	mediaProgress := make([]interface{}, 0)
	if items, ok := stats["recentSessions"].([]interface{}); ok {
		mediaProgress = items
	}

	return mediaProgress, nil
}

// GetCurrentUser 获取当前用户信息
func (c *AbsClient) GetCurrentUser() (*models.AbsUserInfo, error) {
	data, err := c.doRequest("GET", "/api/me", nil)
	if err != nil {
		return nil, err
	}

	var user models.AbsUserInfo
	err = json.Unmarshal(data, &user)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling current user: %w", err)
	}

	return &user, nil
}

// GetListeningStats 获取当前用户的收听统计信息
func (c *AbsClient) GetListeningStats() (map[string]interface{}, error) {
	data, err := c.doRequest("GET", "/api/me/listening-stats", nil)
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
