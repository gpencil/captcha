package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gpencil/captcha"
)

type Handlers struct {
	captchaService *captcha.Service
}

func NewHandlers(service *captcha.Service) *Handlers {
	return &Handlers{
		captchaService: service,
	}
}

// GenerateRequest 生成验证码请求
type GenerateRequest struct {
	CaptchaType string `json:"captchaType"` // character, image_select, slide
}

// GenerateResponse 生成验证码响应
type GenerateResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// VerifyRequest 验证验证码请求
type VerifyRequest struct {
	CaptchaID     string      `json:"captchaId"`
	CaptchaType   string      `json:"captchaType"`
	CaptchaCode   string      `json:"captchaCode,omitempty"`   // 字符验证码
	CaptchaAnswer interface{} `json:"captchaAnswer,omitempty"` // 其他类型验证码答案
}

// VerifyResponse 验证响应
type VerifyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}

// IndexPage 首页
func (h *Handlers) IndexPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/index.html"))
	tmpl.Execute(w, nil)
}

// GenerateCaptcha 生成验证码
func (h *Handlers) GenerateCaptcha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "无效的请求参数", 400)
		return
	}

	ctx := context.Background()
	captchaType := captcha.CaptchaType(req.CaptchaType)

	// 生成验证码
	resp, err := h.captchaService.Generate(ctx, captchaType)
	if err != nil {
		respondWithError(w, "生成验证码失败: "+err.Error(), 500)
		return
	}

	// 保存验证码ID到session（这里使用cookie简化）
	http.SetCookie(w, &http.Cookie{
		Name:     "captcha_id",
		Value:    resp.CaptchaID,
		Expires:  time.Now().Add(5 * time.Minute),
		Path:     "/",
		HttpOnly: true,
	})

	respondWithSuccess(w, resp)
}

// VerifyCaptcha 验证验证码
func (h *Handlers) VerifyCaptcha(w http.ResponseWriter, r *http.Request) {
	fmt.Println("VerifyCaptcha-1")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("VerifyCaptcha-2")
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "无效的请求参数", 400)
		return
	}
	fmt.Println("VerifyCaptcha-3")
	ctx := context.Background()

	// 构建验证请求
	verifyReq := &captcha.VerifyRequest{
		CaptchaID:   req.CaptchaID,
		CaptchaType: captcha.CaptchaType(req.CaptchaType),
	}
	fmt.Println("VerifyCaptcha-4")
	// 根据验证码类型设置答案
	switch req.CaptchaType {
	case "character":
		verifyReq.Answer = captcha.CharacterAnswer{
			Code: req.CaptchaCode,
		}
	case "image_select":
		// 从 CaptchaAnswer 中解析 SelectedIndexes
		if answerMap, ok := req.CaptchaAnswer.(map[string]interface{}); ok {
			selectedIndexes := make([]int, 0)
			if indexes, ok := answerMap["selectedIndexes"].([]interface{}); ok {
				for _, idx := range indexes {
					if i, ok := idx.(float64); ok {
						selectedIndexes = append(selectedIndexes, int(i))
					}
				}
			}
			verifyReq.Answer = captcha.ImageSelectAnswer{
				SelectedIndexes: selectedIndexes,
			}
		}
	case "slide":
		// 从 CaptchaAnswer 中解析滑动数据
		if answerMap, ok := req.CaptchaAnswer.(map[string]interface{}); ok {
			x := int(answerMap["x"].(float64))
			duration := int64(answerMap["duration"].(float64))
			track := make([]int, 0)
			if trackData, ok := answerMap["track"].([]interface{}); ok {
				for _, t := range trackData {
					if v, ok := t.(float64); ok {
						track = append(track, int(v))
					}
				}
			}
			verifyReq.Answer = captcha.SlideAnswer{
				X:        x,
				Track:    track,
				Duration: duration,
			}
		}
	}
	fmt.Println("VerifyCaptcha-5")
	// 验证
	valid, err := h.captchaService.VerifyAndDelete(ctx, verifyReq)
	if err != nil {
		fmt.Println("VerifyCaptcha-6")
		respondWithError(w, "验证失败: "+err.Error(), 500)
		return
	}
	fmt.Println("VerifyCaptcha-7", valid)
	respondWithVerifyResult(w, valid)
}

// respondWithSuccess 返回成功响应
func respondWithSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenerateResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// respondWithVerifyResult 返回验证结果
func respondWithVerifyResult(w http.ResponseWriter, valid bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(VerifyResponse{
		Code:    0,
		Message: "success",
		Valid:   valid,
	})
}

// respondWithError 返回错误响应
func respondWithError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	})
}

// 辅助函数：生成UUID
func generateUUID() string {
	return uuid.New().String()
}
