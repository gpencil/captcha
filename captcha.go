package captcha

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/slide"
)

// Service 验证码服务
type Service struct {
	redis   *redis.Client
	builder slide.Builder
	config  Config
}

// Config 配置
type Config struct {
	Width          int
	Height         int
	TemplateWidth  int
	TemplateHeight int
	ExpireTime     time.Duration
	RedisPrefix    string
}

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID       string `json:"captchaId"`
	BackgroundImage string `json:"backgroundImage"`
	TileImage       string `json:"tileImage"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	ExpireTime      int64  `json:"expireTime"`
}

// VerifyRequest 验证请求
type VerifyRequest struct {
	CaptchaID string `json:"captchaId"`
	X         int    `json:"x"`        // 滑块X轴坐标（百分比 0-100）
	Duration  int64  `json:"duration"` // 滑动耗时（毫秒）
}

// NewService 创建服务
func NewService(redisClient *redis.Client, config Config) *Service {
	// 设置默认值
	if config.Width == 0 {
		config.Width = 350
	}
	if config.Height == 0 {
		config.Height = 200
	}
	if config.TemplateWidth == 0 {
		config.TemplateWidth = 60
	}
	if config.TemplateHeight == 0 {
		config.TemplateHeight = 60
	}
	if config.ExpireTime == 0 {
		config.ExpireTime = 5 * time.Minute
	}
	if config.RedisPrefix == "" {
		config.RedisPrefix = "captcha:"
	}

	// 创建背景图片
	backgrounds := make([]image.Image, 0)
	for i := 0; i < 5; i++ {
		bg := createBackground(config.Width, config.Height, i)
		backgrounds = append(backgrounds, bg)
	}

	// 创建滑块图形
	graphImages := make([]*slide.GraphImage, 0)
	for i := 0; i < 3; i++ {
		graph := createGraphImage(config.TemplateWidth, config.TemplateHeight, i)
		graphImages = append(graphImages, graph)
	}

	// 创建构建器
	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: config.Width, Height: config.Height}),
		slide.WithRangeGraphSize(option.RangeVal{Min: config.TemplateWidth, Max: config.TemplateWidth}),
	)
	builder.SetResources(
		slide.WithBackgrounds(backgrounds),
		slide.WithGraphImages(graphImages),
	)

	return &Service{
		redis:   redisClient,
		builder: builder,
		config:  config,
	}
}

// Generate 生成验证码
func (s *Service) Generate(ctx context.Context) (*CaptchaResponse, error) {
	// 生成验证码
	capt := s.builder.Make()
	captData, err := capt.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate captcha: %w", err)
	}

	// 获取主图和滑块图
	masterImage := captData.GetMasterImage()
	tileImage := captData.GetTileImage()

	// 转换为Base64
	masterBase64, err := imageToBase64(masterImage.Get())
	if err != nil {
		return nil, fmt.Errorf("failed to encode master image: %w", err)
	}

	tileBase64, err := imageToBase64(tileImage.Get())
	if err != nil {
		return nil, fmt.Errorf("failed to encode tile image: %w", err)
	}

	// 获取正确的X坐标
	block := captData.GetData()
	targetX := block.X

	// 生成验证码ID
	captchaID := uuid.New().String()

	// 存储到Redis
	data := map[string]interface{}{
		"targetX": targetX,
	}
	dataBytes, _ := json.Marshal(data)
	key := s.config.RedisPrefix + captchaID
	err = s.redis.Set(ctx, key, dataBytes, s.config.ExpireTime).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store captcha: %w", err)
	}

	return &CaptchaResponse{
		CaptchaID:       captchaID,
		BackgroundImage: masterBase64,
		TileImage:       tileBase64,
		Width:           s.config.Width,
		Height:          s.config.Height,
		ExpireTime:      time.Now().Add(s.config.ExpireTime).Unix(),
	}, nil
}

// Verify 验证验证码
func (s *Service) Verify(ctx context.Context, req *VerifyRequest) (bool, error) {
	// 从Redis获取数据
	key := s.config.RedisPrefix + req.CaptchaID
	dataBytes, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get captcha: %w", err)
	}

	// 解析数据
	var data map[string]interface{}
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return false, err
	}

	targetX := int(data["targetX"].(float64))

	// 将百分比转换为实际像素位置
	actualX := int(float64(req.X) * float64(s.config.Width-s.config.TemplateWidth) / 100.0)

	// 允许误差范围
	tolerance := 10
	diff := math.Abs(float64(targetX - actualX))
	if diff > float64(tolerance) {
		return false, nil
	}

	// 检查滑动时间
	minDuration := int64(300)   // 0.3秒
	maxDuration := int64(30000) // 30秒
	if req.Duration < minDuration || req.Duration > maxDuration {
		return false, nil
	}

	// 验证成功，删除验证码
	s.redis.Del(ctx, key)

	return true, nil
}

// imageToBase64 将图片编码为Base64
func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// createBackground 创建背景图
func createBackground(width, height, seed int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 创建渐变背景
	colors := []struct{ r, g, b uint8 }{
		{100, 150, 200},
		{120, 170, 220},
		{140, 190, 240},
		{80, 130, 180},
		{110, 160, 210},
	}
	c := colors[seed%len(colors)]

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := c.r + uint8(x*100/width)
			g := c.g + uint8(y*50/height)
			b := c.b
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// createGraphImage 创建拼图形状
func createGraphImage(width, height, seed int) *slide.GraphImage {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	mask := image.NewRGBA(image.Rect(0, 0, width, height))

	centerX := width / 2
	centerY := height / 2
	radius := width / 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := x - centerX
			dy := y - centerY
			distance := math.Sqrt(float64(dx*dx + dy*dy))

			// 创建圆形拼图块
			if distance <= float64(radius) {
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
				mask.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			} else {
				img.Set(x, y, color.RGBA{A: 0})
				mask.Set(x, y, color.RGBA{A: 0})
			}
		}
	}

	return &slide.GraphImage{
		OverlayImage: img,
		ShadowImage:  img,
		MaskImage:    mask,
	}
}
