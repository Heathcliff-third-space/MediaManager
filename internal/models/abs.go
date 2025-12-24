package models

// AbsServerStatus 服务器状态信息
type AbsServerStatus struct {
	Success       bool   `json:"success"`
	ServerVersion string `json:"serverVersion"`
	APIVersion    string `json:"apiVersion"`
	UserID        string `json:"userId,omitempty"`
	Username      string `json:"username,omitempty"`
	Language      string `json:"language,omitempty"`
}

// AbsLibraryInfo 媒体库信息
type AbsLibraryInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Folders []struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	} `json:"folders"`
	DisplayOrder int64       `json:"displayOrder"`
	Icon         string      `json:"icon"`
	LastScan     int64       `json:"lastScan,omitempty"` // 修改为int64
	CreatedAt    int64       `json:"createdAt"`
	UpdatedAt    int64       `json:"updatedAt"`
	MediaType    string      `json:"mediaType"`
	Provider     string      `json:"provider"`
	Settings     interface{} `json:"settings"`
	ItemCount    int         `json:"itemCount,omitempty"`
}

// AbsBook 书籍信息
// 根据API响应，我们只需要用到path、size、addedAt字段
// 这些字段来自libraryItem对象
// 现在添加libraryId字段以显示对应的媒体库，并使用relPath代替path以提高安全性
type AbsBook struct {
	LibraryID string `json:"libraryId"`
	RelPath   string `json:"relPath"`
	Size      int64  `json:"size"`
	AddedAt   int64  `json:"addedAt"`
}

// AbsServerInfo 服务器基本信息
type AbsServerInfo struct {
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
}

// AbsUserInfo 用户信息
type AbsUserInfo struct {
	ID            string        `json:"id"`
	Username      string        `json:"username"`
	Type          string        `json:"type"`
	Token         string        `json:"token,omitempty"`
	IsActive      bool          `json:"isActive"`
	LastSeen      int64         `json:"lastSeen"`
	MediaProgress []interface{} `json:"mediaProgress"` // 根据实际情况调整类型
	CreatedAt     int64         `json:"createdAt"`
	UpdatedAt     int64         `json:"updatedAt"`
}
