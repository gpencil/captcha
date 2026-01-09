package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// CharacterCaptcha 字符验证码
type CharacterCaptcha struct {
	config CharacterConfig
}

// NewCharacterCaptcha 创建字符验证码
func NewCharacterCaptcha(config CharacterConfig) *CharacterCaptcha {
	if config.Width == 0 {
		config.Width = 160
	}
	if config.Height == 0 {
		config.Height = 60
	}
	if config.Length == 0 {
		config.Length = 4
	}
	if config.ExpireTime == 0 {
		config.ExpireTime = 5 * time.Minute
	}
	if config.Complexity == 0 {
		config.Complexity = 2
	}

	return &CharacterCaptcha{
		config: config,
	}
}

// Generate 生成验证码
func (c *CharacterCaptcha) Generate() (string, string, error) {
	// 生成随机验证码
	code := c.generateCode()

	// 生成图片
	imageBytes, err := c.generateImage(code)
	if err != nil {
		return "", "", err
	}

	// 转换为 Base64
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)
	// 添加 data URI 前缀
	base64Image = "data:image/png;base64," + base64Image

	return code, base64Image, nil
}

// Verify 验证验证码
func (c *CharacterCaptcha) Verify(code, answer string) bool {
	fmt.Println("验证验证码:", code, "===", answer)
	// 不区分大小写比较
	return strings.EqualFold(code, answer)
}

// generateCode 生成随机验证码
func (c *CharacterCaptcha) generateCode() string {
	charset := "23456789ABCDEFGHKMNPRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	code := make([]byte, c.config.Length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}

	return string(code)
}

// generateImage 生成验证码图片
func (c *CharacterCaptcha) generateImage(code string) ([]byte, error) {
	// 创建 RGBA 图片
	img := image.NewRGBA(image.Rect(0, 0, c.config.Width, c.config.Height))

	// 填充白色背景
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)

	// 添加背景噪点
	c.addNoise(img)

	// 添加干扰线
	c.addLines(img)

	// 添加文字
	c.drawText(img, code)

	// 编码为 PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// drawText 绘制文字
func (c *CharacterCaptcha) drawText(img *image.RGBA, code string) {
	// 使用基础字体
	face := basicfont.Face7x13

	// 计算缩放因子（放大 2.5 倍）
	scale := 2.5

	// 计算每个字符的宽度和位置
	charWidth := (c.config.Width - 20) / c.config.Length
	charHeight := c.config.Height / 4 // 调整为 1/4，让字体位置更靠上，/3往下， /5往上

	rand.Seed(time.Now().UnixNano())

	for i, ch := range code {
		// 随机位置
		x := 10 + i*charWidth + rand.Intn(10)
		y := charHeight + rand.Intn(5)

		// 随机颜色
		textColor := color.RGBA{
			R: uint8(rand.Intn(128)),
			G: uint8(rand.Intn(128)),
			B: uint8(rand.Intn(128)),
			A: 255,
		}

		// 先在小的临时图片上绘制原始字符
		tmpWidth := 8
		tmpHeight := 13
		tmpImg := image.NewRGBA(image.Rect(0, 0, tmpWidth, tmpHeight))

		d := font.Drawer{
			Dst:  tmpImg,
			Src:  &image.Uniform{textColor},
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(12)},
		}
		d.DrawString(string(ch))

		// 将临时图片的像素放大绘制到目标图片
		for ty := 0; ty < tmpHeight; ty++ {
			for tx := 0; tx < tmpWidth; tx++ {
				_, _, _, a := tmpImg.At(tx, ty).RGBA()
				if a > 0 {
					// 放大绘制：每个原始像素变成 scale x scale 的方块
					for sy := 0; sy < int(scale); sy++ {
						for sx := 0; sx < int(scale); sx++ {
							targetX := x + int(float64(tx)*scale) + sx
							targetY := y + int(float64(ty)*scale) + sy
							if targetX < c.config.Width && targetY < c.config.Height && targetY >= 0 {
								img.Set(targetX, targetY, textColor)
							}
						}
					}
				}
			}
		}
	}
}

// addNoise 添加噪点
func (c *CharacterCaptcha) addNoise(img *image.RGBA) {
	rand.Seed(time.Now().UnixNano())
	noiseDensity := 50 * c.config.Complexity

	for i := 0; i < noiseDensity; i++ {
		x := rand.Intn(c.config.Width)
		y := rand.Intn(c.config.Height)

		img.Set(x, y, color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 255,
		})
	}
}

// addLines 添加干扰线
func (c *CharacterCaptcha) addLines(img *image.RGBA) {
	rand.Seed(time.Now().UnixNano())
	lineCount := c.config.Complexity * 2

	for i := 0; i < lineCount; i++ {
		x1 := rand.Intn(c.config.Width)
		y1 := rand.Intn(c.config.Height)
		x2 := rand.Intn(c.config.Width)
		y2 := rand.Intn(c.config.Height)

		lineColor := color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 255,
		}

		c.drawLine(img, x1, y1, x2, y2, lineColor)
	}
}

// drawLine 绘制线条
func (c *CharacterCaptcha) drawLine(img *image.RGBA, x1, y1, x2, y2 int, lineColor color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < c.config.Width && y1 >= 0 && y1 < c.config.Height {
			img.Set(x1, y1, lineColor)
		}

		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// init 初始化
func init() {
	logx.Info("Character captcha initialized")
}
