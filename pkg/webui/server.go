package webui

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gowsp/cloud189/pkg"
)

// Server Web服务器结构
type Server struct {
	app    pkg.Drive
	engine *gin.Engine
}

// NewServer 创建新的Web服务器
func NewServer(app pkg.Drive) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()
	
	server := &Server{
		app:    app,
		engine: engine,
	}
	
	server.setupRoutes()
	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 静态文件服务
	s.engine.Static("/static", "./web/static")
	s.engine.LoadHTMLGlob("web/templates/*")
	
	// 主页
	s.engine.GET("/", s.handleIndex)
	
	// API路由组
	api := s.engine.Group("/api")
	{
		api.GET("/files", s.handleListFiles)
		api.GET("/files/:id", s.handleGetFile)
		api.POST("/files/upload", s.handleUpload)
		api.GET("/files/:id/download", s.handleDownload)
		api.POST("/files/mkdir", s.handleMkdir)
		api.DELETE("/files/:id", s.handleDelete)
		api.PUT("/files/:id/rename", s.handleRename)
		api.POST("/files/move", s.handleMove)
		api.GET("/search", s.handleSearch)
		api.GET("/space", s.handleSpace)
	}
}

// Start 启动Web服务器
func (s *Server) Start(addr string) error {
	return s.engine.Run(addr)
}

// handleIndex 处理主页请求
func (s *Server) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "天翼云盘 Web 管理界面",
	})
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// success 成功响应
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// error 错误响应
func errorResponse(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// FileInfo 文件信息结构
type FileInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"isDir"`
	ModTime  string `json:"modTime"`
}

// convertFileInfo 转换文件信息
func convertFileInfo(file pkg.File) FileInfo {
	return FileInfo{
		ID:      file.Id(),
		Name:    file.Name(),
		Size:    file.Size(),
		IsDir:   file.IsDir(),
		ModTime: file.ModTime().Format("2006-01-02 15:04:05"),
	}
}