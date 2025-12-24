package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Heathcliff-third-space/MediaManager/internal/config"
	"io"
	"net/http"
	"net/url"
	"time"
)

// EmbyClient represents an Emby API client
type EmbyClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewEmbyClient creates a new Emby API client
func NewEmbyClient(config *config.Config) *EmbyClient {
	baseURL := config.EmbyURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", config.EmbyPort)
	}

	// 创建不使用代理的 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &EmbyClient{
		baseURL:    baseURL,
		apiKey:     config.EmbyToken,
		httpClient: client,
	}
}

// doRequest performs an HTTP request to the Emby API
func (c *EmbyClient) doRequest(method, path string, body interface{}) ([]byte, error) {
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

	// Emby API 使用 API Key 认证
	req.Header.Set("X-Emby-Token", c.apiKey)
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

// GetSystemInfo 获取 Emby 服务器信息
func (c *EmbyClient) GetSystemInfo() ([]byte, error) {
	return c.doRequest("GET", "/System/Info", nil)
}

// GetUsers 获取用户列表
func (c *EmbyClient) GetUsers() ([]byte, error) {
	return c.doRequest("GET", "/Users", nil)
}

// GetCurrentUser 获取当前用户信息
func (c *EmbyClient) GetCurrentUser() ([]byte, error) {
	return c.doRequest("GET", "/Users/Me", nil)
}

// GetMediaFolders 获取媒体库（媒体文件夹）
func (c *EmbyClient) GetMediaFolders() ([]byte, error) {
	return c.doRequest("GET", "/Library/MediaFolders", nil)
}

// GetItems 获取媒体项目
func (c *EmbyClient) GetItems(parentID string, userID string, itemTypes []string) ([]byte, error) {
	params := url.Values{}
	if parentID != "" {
		params.Add("ParentId", parentID)
	}
	if userID != "" {
		params.Add("UserId", userID)
	}
	if len(itemTypes) > 0 {
		params.Add("IncludeItemTypes", fmt.Sprintf("%s", itemTypes[0]))
		for _, t := range itemTypes[1:] {
			params.Add("IncludeItemTypes", params.Get("IncludeItemTypes")+","+t)
		}
	}
	params.Add("Recursive", "true")

	path := "/Items"
	if params.Encode() != "" {
		path += "?" + params.Encode()
	}

	return c.doRequest("GET", path, nil)
}

// SearchItems 搜索媒体项目
func (c *EmbyClient) SearchItems(searchTerm string, userID string, limit int) ([]byte, error) {
	params := url.Values{}
	params.Add("SearchTerm", searchTerm)
	// 使用更全面的搜索端点
	params.Add("IncludeItemTypes", "Movie,Series,MusicAlbum,MusicArtist,Playlist,Audio,Book,Folder,Photo,PhotoAlbum")
	// 添加必要的字段参数以获取完整媒体信息
	params.Add("Fields", "Path,DateCreated,Size,Overview,ProviderIds,Genres,Studios,Taglines,LocalTrailerCount,OfficialRating,CumulativeRunTimeTicks,ItemCounts,DisplayPreferencesId,ChildCount,RecursiveChildCount,ProductionLocations,CriticRating,ShortOverview,MediaSourceCount,PrimaryImageAspectRatio")
	// 包括所有类型的媒体
	params.Add("IncludePeople", "true")
	params.Add("IncludeGenres", "true")
	params.Add("IncludeStudios", "true")
	params.Add("IncludeArtists", "true")
	// 排除剧集类型，确保搜索结果中不包含剧集类型的媒体
	params.Add("ExcludeItemTypes", "Episode")
	params.Add("Recursive", "true")
	params.Add("EnableTotalRecordCount", "false")
	if userID != "" {
		params.Add("UserId", userID)
	}
	if limit > 0 {
		params.Add("Limit", fmt.Sprintf("%d", limit))
	}

	path := "/Items" + "?" + params.Encode()

	return c.doRequest("GET", path, nil)
}

// GetItemCounts 获取项目统计信息
func (c *EmbyClient) GetItemCounts() ([]byte, error) {
	return c.doRequest("GET", "/Items/Counts", nil)
}

// GetLibraryItemsCount 获取指定媒体库的项目数量
func (c *EmbyClient) GetLibraryItemsCount(libraryID string) (int, error) {
	// 使用Items端点获取媒体库项目数量
	params := url.Values{}
	params.Add("ParentId", libraryID)
	params.Add("Recursive", "true")
	params.Add("EnableTotalRecordCount", "true")
	
	path := "/Items" + "?" + params.Encode()
	
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}
	
	var response struct {
		Items            []interface{} `json:"Items"`
		TotalRecordCount int           `json:"TotalRecordCount"`
	}
	
	err = json.Unmarshal(data, &response)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling library items count: %w", err)
	}
	
	return response.TotalRecordCount, nil
}

// GetUserData 获取用户的媒体播放进度
func (c *EmbyClient) GetUserData(userID string) ([]byte, error) {
	return c.doRequest("GET", fmt.Sprintf("/Users/%s/Items", userID), nil)
}

// GetResumeItems 获取用户继续播放的项目
func (c *EmbyClient) GetResumeItems(userID string) ([]byte, error) {
	return c.doRequest("GET", fmt.Sprintf("/Users/%s/Items/Resume", userID), nil)
}
