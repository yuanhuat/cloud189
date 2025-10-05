package webui

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gowsp/cloud189/pkg"
	"github.com/gowsp/cloud189/pkg/file"
)

// handleListFiles 处理文件列表请求
func (s *Server) handleListFiles(c *gin.Context) {
	path := c.DefaultQuery("path", "/")
	
	entries, err := s.app.ReadDir(path)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("获取文件列表失败: %v", err))
		return
	}
	
	var fileInfos []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		fileInfo := FileInfo{
			ID:      entry.Name(), // 使用名称作为ID
			Name:    entry.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		}
		fileInfos = append(fileInfos, fileInfo)
	}
	
	success(c, gin.H{
		"files": fileInfos,
		"path":  path,
	})
}

// handleGetFile 处理获取单个文件信息
func (s *Server) handleGetFile(c *gin.Context) {
	id := c.Param("id")
	
	file, err := s.app.Stat(id)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("获取文件信息失败: %v", err))
		return
	}
	
	// 由于file是fs.FileInfo类型，我们需要创建一个简单的FileInfo结构
	fileInfo := FileInfo{
		ID:      id,
		Name:    file.Name(),
		Size:    file.Size(),
		IsDir:   file.IsDir(),
		ModTime: file.ModTime().Format("2006-01-02 15:04:05"),
	}
	success(c, fileInfo)
}

// handleUpload 处理文件上传
func (s *Server) handleUpload(c *gin.Context) {
	parentPath := c.PostForm("path")
	if parentPath == "" {
		parentPath = "/"
	}
	
	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("获取上传文件失败: %v", err))
		return
	}
	
	// 打开文件
	src, err := fileHeader.Open()
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("打开上传文件失败: %v", err))
		return
	}
	defer src.Close()
	
	// 获取父目录信息
	parent, err := s.app.Stat(parentPath)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("获取父目录失败: %v", err))
		return
	}
	
	// 创建HTTP请求对象用于上传
	req := &http.Request{
		ContentLength: fileHeader.Size,
		Body:         src,
	}
	
	// 创建上传文件对象
	uploadFile := file.NewWebFile(parent.(pkg.File).Id(), fileHeader.Filename, req)
	
	// 执行上传
	err = s.app.UploadFrom(uploadFile)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("上传文件失败: %v", err))
		return
	}
	
	success(c, gin.H{
		"filename": fileHeader.Filename,
		"size":     fileHeader.Size,
	})
}

// handleDownload 处理文件下载请求
func (s *Server) handleDownload(c *gin.Context) {
	fileId := c.Param("id")
	
	// 获取下载链接
	downloadUrl, err := s.app.GetDownloadUrl(fileId)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("获取下载链接失败: %v", err))
		return
	}
	
	// 重定向到下载链接
	c.Redirect(http.StatusFound, downloadUrl)
}

// handleMkdir 处理创建文件夹请求
func (s *Server) handleMkdir(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 1, fmt.Sprintf("参数错误: %v", err))
		return
	}
	
	// 构建完整路径
	fullPath := req.Path
	if fullPath != "/" {
		fullPath += "/"
	}
	fullPath += req.Name
	
	// 创建文件夹
	err := s.app.Mkdir(fullPath)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("创建文件夹失败: %v", err))
		return
	}
	
	success(c, gin.H{
		"name": req.Name,
		"path": req.Path,
	})
}

// handleDelete 处理删除文件请求
func (s *Server) handleDelete(c *gin.Context) {
	fileId := c.Param("id")
	
	// 删除文件
	err := s.app.Delete(fileId)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("删除文件失败: %v", err))
		return
	}
	
	success(c, gin.H{
		"id": fileId,
	})
}

// handleRename 重命名文件
func (s *Server) handleRename(c *gin.Context) {
	// 暂时返回未实现错误，因为Drive接口目前没有Rename方法
	errorResponse(c, http.StatusNotImplemented, "重命名功能暂未实现")
	return
}

// handleMove 处理移动文件请求
func (s *Server) handleMove(c *gin.Context) {
	var req struct {
		SourcePath string `json:"sourcePath" binding:"required"`
		TargetPath string `json:"targetPath" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 1, fmt.Sprintf("参数错误: %v", err))
		return
	}
	
	// 移动文件
	err := s.app.Move(req.TargetPath, req.SourcePath)
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("移动文件失败: %v", err))
		return
	}
	
	success(c, gin.H{
		"sourcePath": req.SourcePath,
		"targetPath": req.TargetPath,
	})
}

// handleSearch 处理搜索请求
func (s *Server) handleSearch(c *gin.Context) {
	keyword := c.Query("keyword")
	
	if keyword == "" {
		errorResponse(c, 1, "搜索关键词不能为空")
		return
	}
	
	// TODO: 实现搜索功能
	// 目前Drive接口没有Search方法，需要后续实现
	errorResponse(c, 1, "搜索功能暂未实现")
}

// handleSpace 处理获取空间信息
func (s *Server) handleSpace(c *gin.Context) {
	space, err := s.app.Space()
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("获取空间信息失败: %v", err))
		return
	}
	
	success(c, gin.H{
		"total": space.Capacity,
		"used":  space.Capacity - space.Available,
		"free":  space.Available,
	})
}

// getFileType 获取文件类型
func getFileType(name string, isDir bool) string {
	if isDir {
		return "folder"
	}
	
	ext := filepath.Ext(name)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return "image"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac":
		return "audio"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "word"
	case ".xls", ".xlsx":
		return "excel"
	case ".ppt", ".pptx":
		return "powerpoint"
	case ".txt":
		return "text"
	case ".zip", ".rar", ".7z":
		return "archive"
	default:
		return "file"
	}
}