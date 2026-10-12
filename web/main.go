package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gpencil/captcha"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取当前工作目录:", err)
	}

	// 检查 static 目录是否存在，如果不存在则尝试切换到 web 目录
	staticDir := filepath.Join(currentDir, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		webDir := filepath.Join(currentDir, "web")
		if _, err := os.Stat(filepath.Join(webDir, "static")); err == nil {
			if err := os.Chdir(webDir); err != nil {
				log.Fatal("无法切换到 web 目录:", err)
			}
			currentDir = webDir
		}
	}

	log.Printf("工作目录: %s", currentDir)

	// 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       0,
	})

	// 创建验证码服务
	service := captcha.NewService(redisClient, captcha.Config{
		Width:          350,
		Height:         200,
		TemplateWidth:  60,
		TemplateHeight: 60,
		ExpireTime:     5 * time.Minute,
	})

	// 设置路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/captcha/generate", generateHandler(service))
	http.HandleFunc("/api/captcha/verify", verifyHandler(service))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 启动服务器
	addr := ":8084"
	log.Printf("验证码测试服务启动成功！访问地址: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/index.html"))
	tmpl.Execute(w, nil)
}

func generateHandler(service *captcha.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := context.Background()
		resp, err := service.Generate(ctx)
		if err != nil {
			respondError(w, "生成验证码失败: "+err.Error(), 500)
			return
		}

		respondSuccess(w, resp)
	}
}

func verifyHandler(service *captcha.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req captcha.VerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, "无效的请求参数", 400)
			return
		}

		ctx := context.Background()
		valid, err := service.Verify(ctx, &req)
		if err != nil {
			respondError(w, "验证失败: "+err.Error(), 500)
			return
		}

		respondSuccess(w, map[string]interface{}{
			"valid": valid,
		})
	}
}

func respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	})
}
