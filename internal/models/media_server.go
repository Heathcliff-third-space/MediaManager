package models

// MediaServer 定义通用媒体服务器接口
type MediaServer interface {
	// GetServerInfo 获取服务器信息
	GetServerInfo() (*ServerInfo, error)
	
	// GetUsers 获取用户列表
	GetUsers() ([]UserInfo, error)
	
	// GetCurrentUser 获取当前用户信息
	GetCurrentUser() (*UserInfo, error)
	
	// GetLibraries 获取媒体库列表
	GetLibraries() ([]LibraryInfo, error)
	
	// GetLibraryItemsCount 获取指定媒体库中的项目数量
	GetLibraryItemsCount(libraryID string) (int, error)
	
	// Search 搜索媒体内容
	Search(query string) ([]SearchResult, error)
	
	// GetListeningStats 获取当前用户的收听/观看统计
	GetListeningStats() (map[string]interface{}, error)
}

// ServerInfo 服务器信息
type ServerInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	PublicIP      string `json:"publicIP"`
	LocalIP       string `json:"localIP"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	StartTime     int64  `json:"startTime"`
	Uptime        int64  `json:"uptime"`
	TotalRAM      int64  `json:"totalRAM"`
	FreeRAM       int64  `json:"freeRAM"`
	TotalDiskSize int64  `json:"totalDiskSize"`
	FreeDiskSize  int64  `json:"freeDiskSize"`
	ServerVersion string `json:"serverVersion"`
	APIVersion    string `json:"apiVersion"`
	Language      string `json:"language"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID            string      `json:"id"`
	Username      string      `json:"username"`
	Type          string      `json:"type"`
	Token         string      `json:"token,omitempty"`
	IsActive      bool        `json:"isActive"`
	LastSeen      int64       `json:"lastSeen"`
	MediaProgress interface{} `json:"mediaProgress"` // 根据实际情况调整类型
	CreatedAt     int64       `json:"createdAt"`
	UpdatedAt     int64       `json:"updatedAt"`
}

// LibraryInfo 媒体库信息
type LibraryInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ItemCount   int    `json:"itemCount"`
	MediaType   string `json:"mediaType"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	LastScan    int64  `json:"lastScan,omitempty"`
}

// SearchResult 搜索结果
type SearchResult struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Size        int64    `json:"size"`
	AddedAt     int64    `json:"addedAt"`
	LibraryID   string   `json:"libraryId"`
	Library     string   `json:"library"`
	Type        string   `json:"type"` // book, movie, series, etc.
	Path        string   `json:"path"`
	RelPath     string   `json:"relPath"`
	Overview    string   `json:"overview,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Year        int      `json:"year,omitempty"`
	ProductionYear int   `json:"productionYear,omitempty"`
	PremiereDate string `json:"premiereDate,omitempty"`
	RunTime     int64    `json:"runTime,omitempty"`
	MediaType   string   `json:"mediaType,omitempty"`
}