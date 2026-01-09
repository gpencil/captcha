package captcha

import "time"

// CaptchaType 验证码类型
type CaptchaType string

const (
	CaptchaTypeCharacter   CaptchaType = "character"    // 字符验证码
	CaptchaTypeImageSelect CaptchaType = "image_select" // 图片选择验证码
	SlideTypeSelect        CaptchaType = "slide"        // 滑动验证码
)

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	// Redis 配置
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// 字符验证码配置
	CharacterConfig CharacterConfig

	// 图片选择验证码配置
	ImageSelectConfig ImageSelectConfig
}

// CharacterConfig 字符验证码配置
type CharacterConfig struct {
	Width      int           // 图片宽度
	Height     int           // 图片高度
	Length     int           // 验证码长度
	ExpireTime time.Duration // 过期时间
	Complexity int           // 复杂度: 1-简单, 2-中等, 3-复杂
}

// ImageSelectConfig 图片选择验证码配置
type ImageSelectConfig struct {
	ImageCount  int           // 选项图片数量
	SelectCount int           // 需要选择的数量
	ExpireTime  time.Duration // 过期时间
	Category    string        // 图片类别: traffic(交通), animal(动物), food(食物)等
	ImageDir    string        // 图片文件目录路径
}

// SlideConfig 滑动验证码配置
type SlideConfig struct {
	Width          int           // 背景图宽度
	Height         int           // 背景图高度
	TemplateWidth  int           // 滑块模板宽度
	TemplateHeight int           // 滑块模板高度
	ExpireTime     time.Duration // 过期时间
	ImageDir       string        // 背景图片目录路径
	TemplateDir    string        // 滑块模板目录路径
}

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID   string      `json:"captchaId"`   // 验证码ID
	CaptchaType CaptchaType `json:"captchaType"` // 验证码类型
	Data        interface{} `json:"data"`        // 验证码数据
	ExpireTime  int64       `json:"expireTime"`  // 过期时间戳
}

// CharacterCaptchaData 字符验证码数据
type CharacterCaptchaData struct {
	Image string `json:"image"` // Base64 编码的图片
}

// ImageSelectCaptchaData 图片选择验证码数据
type ImageSelectCaptchaData struct {
	Question    string   `json:"question"`    // 问题：如"请选择所有的公交车"
	TargetType  string   `json:"targetType"`  // 目标类型
	Images      []string `json:"images"`      // 图片列表（Base64）
	SelectCount int      `json:"selectCount"` // 需要选择的数量
}

// VerifyRequest 验证请求
type VerifyRequest struct {
	CaptchaID   string      `json:"captchaId"`   // 验证码ID
	CaptchaType CaptchaType `json:"captchaType"` // 验证码类型
	Answer      interface{} `json:"answer"`      // 答案
}

// CharacterAnswer 字符验证码答案
type CharacterAnswer struct {
	Code string `json:"code"` // 验证码
}

// ImageSelectAnswer 图片选择验证码答案
type ImageSelectAnswer struct {
	SelectedIndexes []int `json:"selectedIndexes"` // 选中的图片索引
}

// SlideCaptchaData 滑动验证码数据
type SlideCaptchaData struct {
	BackgroundImage string `json:"backgroundImage"` // 背景图片（Base64）
	TemplateImage   string `json:"templateImage"`   // 滑块模板图片（Base64）
	TemplateY       int    `json:"templateY"`       // 滑块在Y轴的位置
	Width           int    `json:"width"`           // 背景图宽度
	Height          int    `json:"height"`          // 背景图高度
}

// SlideAnswer 滑动验证码答案
type SlideAnswer struct {
	X        int   `json:"x"`        // 滑块X轴坐标（百分比 0-100）
	Track    []int `json:"track"`    // 滑动轨迹
	Duration int64 `json:"duration"` // 滑动耗时（毫秒）
}
