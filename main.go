package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	htmlPage = `<html>
<head>
	<title>Everything</title>
</head>
<body>
	%s
</body>
</html>`
)

// GetWindowsDrives 返回 Windows 系统中的所有盘符列表
func GetWindowsDrives() []string {
	var drives []string
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drivePath := string(drive) + ":\\"
		_, err := os.Stat(drivePath)
		if err == nil {
			drives = append(drives, drivePath)
		}
	}
	return drives
}

// ListDirectories 返回指定路径下的所有一级文件夹列表
func ListDirectories(path string) ([]string, error) {
	var dirs []string
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func main() {
	// // 隐藏控制台窗口
	// cmd := exec.Command("cmd", "/c", "start", "cmd", "/c", "dirlook.exe")
	// cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Start()

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r := gin.Default()

	// 处理根路径请求
	r.GET("/", func(c *gin.Context) {
		// 获取所有盘符列表
		drives := GetWindowsDrives()
		html := "<ul>"
		for _, drive := range drives {
			html += fmt.Sprintf(`<li><a href="/%s">%s</a></li>`, drive, drive)
		}
		html += "</ul>"

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf(htmlPage, html))
	})

	// 处理盘符下的一级文件夹请求
	r.GET("/:drive", func(c *gin.Context) {
		drive := c.Param("drive")
		path := filepath.Join(drive, string(filepath.Separator))
		dirs, err := ListDirectories(path)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error: %s", err.Error())
			return
		}
		html := "<ul>"
		for _, dir := range dirs {
			html += fmt.Sprintf(`<li><a href="/%s/%s">%s</a></li>`, drive, dir, dir)
		}
		html += "</ul>"

		// c.HTML(http.StatusOK, "index.html", html)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf(htmlPage, html))
	})

	// 处理文件夹内的文件和文件夹请求
	r.GET("/:drive/*path", func(c *gin.Context) {
		drive := c.Param("drive")
		relPath := c.Param("path")
		path := filepath.Join(drive, relPath)
		dirs, err := ListDirectories(path)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error: %s", err.Error())
			return
		}
		files, err := ioutil.ReadDir(path)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error: %s", err.Error())
			return
		}
		html := "<ul>"
		for _, dir := range dirs {
			html += fmt.Sprintf(`<li><a href="/%s/%s/%s">%s</a></li>`, drive, relPath, dir, dir)
		}
		for _, file := range files {
			if !file.IsDir() {
				html += fmt.Sprintf(`<li><a href="/download/%s/%s">%s</a></li>`, drive, relPath+"/"+file.Name(), file.Name())
			}
		}
		html += "</ul>"

		// c.HTML(http.StatusOK, "index.html", html)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf(htmlPage, html))
	})

	// 处理文件下载请求
	r.GET("/download/:drive/*path", func(c *gin.Context) {
		drive := c.Param("drive")
		relPath := c.Param("path")
		path := filepath.Join(drive, relPath)
		c.File(path)
	})

	// 启动服务器
	srv := &http.Server{
		Addr:    ":834",
		Handler: r,
	}

	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("listen: %s\n", err)
	// 	}
	// }()

	fmt.Println("dir server run at 834 port")
	fmt.Println("auth: tony, telegram: echoty")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	// gracefulExitWeb(srv)
}

// 优雅退出
func gracefulExitWeb(server *http.Server) {
	quit := make(chan os.Signal, 4)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit

	fmt.Println("got a signal\njingpin site stoped", sig)

	now := time.Now()
	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(cxt); err != nil {
		fmt.Println("err", err)
	}

	// 看看实际退出所耗费的时间
	fmt.Println("------exited--------", time.Since(now))
	os.Exit(0)
}
