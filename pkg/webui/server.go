package webui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

	// 设置会话存储
	store := cookie.NewStore([]byte("cloud189-secret-key-change-in-production"))
	store.Options(sessions.Options{
		MaxAge:   int(24 * time.Hour / time.Second), // 24小时
		HttpOnly: true,
		Secure:   false,                // 在生产环境中应设置为true（需要HTTPS）
		SameSite: http.SameSiteLaxMode, // 使用Lax模式支持重定向
		Path:     "/",                  // 确保cookie在整个站点有效
	})
	engine.Use(sessions.Sessions("cloud189-session", store))

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

	// 登录页面（无需认证）
	s.engine.GET("/", s.handleLogin)
	s.engine.GET("/login", s.handleLogin)

	// 认证API（无需认证）
	auth := s.engine.Group("/api/auth")
	{
		auth.POST("/login", s.handleAuthLogin)
		auth.POST("/logout", s.handleAuthLogout)
	}

	// 天翼云API（需要web认证）
	cloud := s.engine.Group("/api/cloud")
	cloud.Use(s.authMiddleware())
	{
		cloud.POST("/login", s.handleCloudLogin)
		cloud.POST("/logout", s.handleCloudLogout)
		cloud.GET("/qr", s.handleCloudQR)
		cloud.GET("/qr/status", s.handleCloudQRStatus)
	}

	// 需要认证的路由
	authenticated := s.engine.Group("/")
	authenticated.Use(s.authMiddleware())
	{
		// 仪表盘页面
		authenticated.GET("/dashboard", s.handleIndex)

		// API路由组
		api := authenticated.Group("/api")
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
}

// Start 启动Web服务器
func (s *Server) Start(addr string) error {
	return s.engine.Run(addr)
}

// handleLogin 处理登录页面请求
func (s *Server) handleLogin(c *gin.Context) {
	// 检查是否已经登录
	session := sessions.Default(c)
	if session.Get("authenticated") == true {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "系统登录 - 天翼云盘 Web 管理",
	})
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
	ID      string `json:"id"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime string `json:"modTime"`
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

// authMiddleware 认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		authenticated := session.Get("authenticated")

		if authenticated != true {
			// 如果是API请求，返回JSON错误
			path := c.FullPath()
			if c.Request.Header.Get("Content-Type") == "application/json" ||
				c.Request.Header.Get("Accept") == "application/json" ||
				(len(path) >= 4 && path[:4] == "/api") {
				c.JSON(http.StatusUnauthorized, Response{
					Code:    401,
					Message: "未授权访问，请先登录",
				})
				c.Abort()
				return
			}

			// 否则重定向到登录页面
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

// handleAuthLogin 处理登录认证
func (s *Server) handleAuthLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 1, "参数错误")
		return
	}

	// 简单的用户名密码验证（在生产环境中应该使用更安全的方式）
	// 这里可以根据需要修改用户名和密码
	if req.Username == "admin" && req.Password == "admin123" {
		session := sessions.Default(c)
		session.Set("authenticated", true)
		session.Set("username", req.Username)

		if err := session.Save(); err != nil {
			errorResponse(c, 1, "会话保存失败")
			return
		}

		success(c, gin.H{
			"message": "登录成功",
		})
	} else {
		errorResponse(c, 1, "用户名或密码错误")
	}
}

// handleAuthLogout 处理登出
func (s *Server) handleAuthLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	success(c, gin.H{
		"message": "登出成功",
	})
}

// handleCloudLogin 处理天翼云密码登录
func (s *Server) handleCloudLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 1, "参数错误")
		return
	}

	// 调用天翼云登录API
	err := s.app.Login(req.Username, req.Password)
	if err != nil {
		errorResponse(c, 1, "登录失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"message": "天翼云账号登录成功",
	})
}

// handleCloudLogout 处理天翼云退出
func (s *Server) handleCloudLogout(c *gin.Context) {
	// 调用app的Logout方法清除天翼云session和配置
	err := s.app.Logout()
	if err != nil {
		errorResponse(c, 1, fmt.Sprintf("退出失败: %v", err))
		return
	}

	success(c, gin.H{
		"message": "天翼云账号退出成功",
	})
}

// QRCodeResponse 二维码响应结构
type QRCodeResponse struct {
	QRCodeURL string `json:"qrCodeUrl"`
	UUID      string `json:"uuid"`
}

// handleCloudQR 处理二维码生成
func (s *Server) handleCloudQR(c *gin.Context) {
	// 这里需要实现二维码登录的逻辑
	// 由于二维码登录比较复杂，暂时返回一个占位符
	errorResponse(c, 1, "二维码登录功能正在开发中")
}

// handleCloudQRStatus 处理二维码状态查询
func (s *Server) handleCloudQRStatus(c *gin.Context) {
	// 这里需要实现二维码状态查询的逻辑
	errorResponse(c, 1, "二维码登录功能正在开发中")
}
