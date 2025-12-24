package util

import "strings"

// GetMediaTypeIcon 根据媒体类型返回相应的图标
func GetMediaTypeIcon(mediaType string) string {
	switch strings.ToLower(mediaType) {
	case "movie", "movies":
		return "🎬" // 电影
	case "series", "episode", "tvshows":
		return "📺" // 电视剧/剧集
	case "music", "audio", "audiobook", "book":
		return "🎧" // 音乐/有声书/书籍
	case "musicalbum":
		return "💿" // 音乐专辑
	case "folder":
		return "📁" // 文件夹
	case "photo", "image":
		return "🖼️" // 照片/图片
	case "podcast":
		return "🎙️" // 播客
	case "boxsets":
		return "📦" // 合集
	default:
		return "🎭" // 默认媒体类型
	}
}
