package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// SlideCaptcha 滑动验证码
type SlideCaptcha struct {
	config SlideConfig
}

// NewSlideCaptcha 创建滑动验证码
func NewSlideCaptcha(config SlideConfig) *SlideCaptcha {
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

	return &SlideCaptcha{
		config: config,
	}
}

// Generate 生成验证码
func (c *SlideCaptcha) Generate() (string, string, string, int, error) {
	rand.Seed(time.Now().UnixNano())

	// 1. 创建背景图
	backgroundImg := c.createBackgroundImage()

	// 2. 创建滑块模板
	templateImg, templateY := c.createTemplateImage()

	// 3. 在背景图上生成缺口（挖空滑块位置）
	captchaX := c.randomPosition(c.config.Width - c.config.TemplateWidth)
	backgroundWithHole := c.cutHole(backgroundImg, captchaX, templateY, c.config.TemplateWidth, c.config.TemplateHeight)

	// 4. 编码为Base64
	backgroundBase64, err := c.encodeImageToBase64(backgroundWithHole)
	if err != nil {
		return "", "", "", 0, err
	}

	templateBase64, err := c.encodeImageToBase64(templateImg)
	if err != nil {
		return "", "", "", 0, err
	}

	return backgroundBase64, templateBase64, "", captchaX, nil
}

// Verify 验证验证码
func (c *SlideCaptcha) Verify(targetX int, answer SlideAnswer) bool {
	// 允许误差范围
	tolerance := 5 // 像素

	// 检查X坐标是否在允许范围内
	diff := math.Abs(float64(targetX - answer.X))
	if diff > float64(tolerance) {
		return false
	}

	// 检查滑动轨迹（简单验证：轨迹点数量）
	if len(answer.Track) < 10 {
		return false
	}

	// 检查滑动时间（简单验证：不能太快也不能太慢）
	minDuration := int64(500)   // 0.5秒
	maxDuration := int64(10000) // 10秒
	if answer.Duration < minDuration || answer.Duration > maxDuration {
		return false
	}

	return true
}

// createBackgroundImage 创建背景图
func (c *SlideCaptcha) createBackgroundImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.config.Width, c.config.Height))

	// 填充渐变背景
	for y := 0; y < c.config.Height; y++ {
		for x := 0; x < c.config.Width; x++ {
			r := uint8(100 + x*255/c.config.Width)
			g := uint8(150 + y*255/c.config.Height)
			b := uint8(200)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// 添加一些随机噪点
	for i := 0; i < 100; i++ {
		x := rand.Intn(c.config.Width)
		y := rand.Intn(c.config.Height)
		img.Set(x, y, color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 255,
		})
	}

	return img
}

// createTemplateImage 创建滑块模板
func (c *SlideCaptcha) createTemplateImage() (*image.RGBA, int) {
	templateY := rand.Intn(c.config.Height-c.config.TemplateHeight-20) + 10
	img := image.NewRGBA(image.Rect(0, 0, c.config.TemplateWidth, c.config.TemplateHeight))

	// 创建半透明的滑块
	for y := 0; y < c.config.TemplateHeight; y++ {
		for x := 0; x < c.config.TemplateWidth; x++ {
			// 创建圆形滑块
			centerX := c.config.TemplateWidth / 2
			centerY := c.config.TemplateHeight / 2
			radius := c.config.TemplateWidth / 2

			dx := x - centerX
			dy := y - centerY
			distance := math.Sqrt(float64(dx*dx + dy*dy))

			if distance <= float64(radius) {
				// 滑块内部：半透明白色
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 200})
			} else {
				// 滑块外部：透明
				img.Set(x, y, color.RGBA{A: 0})
			}
		}
	}

	// 添加边框
	c.addBorder(img, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	return img, templateY
}

// cutHole 在背景图上挖空滑块位置
func (c *SlideCaptcha) cutHole(background *image.RGBA, x, y, width, height int) *image.RGBA {
	result := image.NewRGBA(background.Bounds())
	draw.Draw(result, result.Bounds(), background, image.Point{}, draw.Src)

	// 在指定位置绘制透明区域（模拟缺口）
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			centerX := width / 2
			centerY := height / 2
			radius := width / 2

			dx := px - centerX
			dy := py - centerY
			distance := math.Sqrt(float64(dx*dx + dy*dy))

			if distance <= float64(radius) {
				// 设置为半透明黑色
				if x+px < c.config.Width && y+py < c.config.Height {
					result.Set(x+px, y+py, color.RGBA{R: 50, G: 50, B: 50, A: 150})
				}
			}
		}
	}

	return result
}

// addBorder 添加边框
func (c *SlideCaptcha) addBorder(img *image.RGBA, borderColor color.RGBA) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// 上边
	for x := 0; x < width; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, height-1, borderColor)
	}

	// 左右边
	for y := 0; y < height; y++ {
		img.Set(0, y, borderColor)
		img.Set(width-1, y, borderColor)
	}
}

// randomPosition 生成随机位置
func (c *SlideCaptcha) randomPosition(max int) int {
	// 在30%-70%范围内随机
	min := int(float64(max) * 0.3)
	maxPos := int(float64(max) * 0.7)
	return min + rand.Intn(maxPos-min)
}

// encodeImageToBase64 将图片编码为Base64
func (c *SlideCaptcha) encodeImageToBase64(img *image.RGBA) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// init 初始化
func init() {
	logx.Info("Slide captcha initialized")
}
