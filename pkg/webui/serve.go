package webui

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gowsp/cloud189/pkg"
)

// Serve 启动Web服务器
func Serve(port string, app pkg.Drive) {
	server := NewServer(app)
	
	fmt.Printf("Starting web server on port %s\n", port)
	fmt.Printf("Web interface available at: http://localhost:%s\n", port)
	
	if err := http.ListenAndServe(":"+port, server.engine); err != nil {
		log.Fatal("Failed to start web server:", err)
	}
}