package main

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"

	"github.com/gpencil/captcha"
)

func main() {
	// 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       0,
	})

	// 创建验证码存储
	store := captcha.NewRedisStore(redisClient, "captcha:")

	// 创建验证码服务（支持所有三种类型）
	service := captcha.NewService(
		store,
		captcha.CharacterConfig{
			Width:      160,
			Height:     60,
			Length:     4,
			ExpireTime: 5 * time.Minute, // 5分钟
			Complexity: 2,
		},
		captcha.ImageSelectConfig{
			ImageCount:  4,
			SelectCount: 1,
			ExpireTime:  5 * time.Minute,
			Category:    "traffic",
			ImageDir:    "./images/traffic",
		},
		captcha.SlideConfig{
			Width:          350,
			Height:         200,
			TemplateWidth:  60,
			TemplateHeight: 60,
			ExpireTime:     5 * time.Minute,
			ImageDir:       "./images/backgrounds",
			TemplateDir:    "./images/templates",
		},
	)

	// 创建处理器
	h := NewHandlers(service)

	// 设置路由
	http.HandleFunc("/", h.IndexPage)
	http.HandleFunc("/api/captcha/generate", h.GenerateCaptcha)
	http.HandleFunc("/api/captcha/verify", h.VerifyCaptcha)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 启动服务器
	addr := ":8083"
	log.Printf("验证码测试服务启动成功！访问地址: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
