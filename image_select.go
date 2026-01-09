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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// ImageSelectCaptcha 图片选择验证码
type ImageSelectCaptcha struct {
	config ImageSelectConfig
}

// NewImageSelectCaptcha 创建图片选择验证码
func NewImageSelectCaptcha(config ImageSelectConfig) *ImageSelectCaptcha {
	// 设置默认值
	if config.ImageCount == 0 {
		config.ImageCount = 4 // 默认4张图片
	}
	if config.SelectCount == 0 {
		config.SelectCount = 1 // 默认选择1张
	}
	if config.ExpireTime == 0 {
		config.ExpireTime = 5 * time.Minute
	}

	return &ImageSelectCaptcha{
		config: config,
	}
}

// Generate 生成验证码
func (c *ImageSelectCaptcha) Generate() (string, []string, []int, error) {
	// 1. 随机选择目标图片索引
	targetIndexes := c.selectTargetIndexes()

	// 2. 加载图片并编码为Base64
	images, err := c.loadImages(targetIndexes)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to load images: %w", err)
	}

	// 3. 生成问题文本
	question := c.generateQuestion()

	return question, images, targetIndexes, nil
}

// Verify 验证验证码
func (c *ImageSelectCaptcha) Verify(targetIndexes, selectedIndex []int) bool {
	if len(selectedIndex) != c.config.SelectCount {
		return false
	}

	// 检查选中的索引是否都在目标索引中
	selectedMap := make(map[int]bool)
	for _, idx := range selectedIndex {
		selectedMap[idx] = true
	}

	for _, targetIdx := range targetIndexes {
		if !selectedMap[targetIdx] {
			return false
		}
	}

	return true
}

// selectTargetIndexes 随机选择目标图片索引
func (c *ImageSelectCaptcha) selectTargetIndexes() []int {
	rand.Seed(time.Now().UnixNano())

	// 创建0到ImageCount-1的索引数组
	allIndexes := make([]int, c.config.ImageCount)
	for i := 0; i < c.config.ImageCount; i++ {
		allIndexes[i] = i
	}

	// 随机打乱
	rand.Shuffle(len(allIndexes), func(i, j int) {
		allIndexes[i], allIndexes[j] = allIndexes[j], allIndexes[i]
	})

	// 取前SelectCount个作为目标
	return allIndexes[:c.config.SelectCount]
}

// loadImages 加载图片
func (c *ImageSelectCaptcha) loadImages(targetIndexes []int) ([]string, error) {
	// 首先尝试从文件系统加载真实图片
	realImages, err := c.loadRealImages()
	if err == nil && len(realImages) > 0 {
		return realImages, nil
	}

	// 如果没有真实图片，使用占位图片（回退方案）
	logx.Infof("未找到真实图片，使用占位图片")
	return c.generatePlaceholderImages()
}

// loadRealImages 从文件系统加载真实图片
func (c *ImageSelectCaptcha) loadRealImages() ([]string, error) {
	var loadedImages []string

	// 定义图片类型和对应的目录
	imageTypes := []struct {
		dirName string
	}{
		{"bus"},
		{"car"},
		{"bike"},
		{"light"},
	}

	// 为每种类型随机选择一张图片
	for _, imgType := range imageTypes {
		// 查找该类型的图片文件
		dirPath := filepath.Join(c.config.ImageDir, imgType.dirName)

		// 查找目录中的所有图片文件
		files, err := os.ReadDir(dirPath)
		if err != nil {
			logx.Errorf("读取目录失败: %s, error: %v", dirPath, err)
			continue // 跳过这个类型，尝试下一个
		}

		// 过滤出图片文件
		var imageFiles []string
		for _, file := range files {
			if !file.IsDir() {
				name := strings.ToLower(file.Name())
				if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") ||
					strings.HasSuffix(name, ".png") {
					imageFiles = append(imageFiles, filepath.Join(dirPath, file.Name()))
				}
			}
		}

		if len(imageFiles) == 0 {
			logx.Infof("目录 %s 中没有找到图片文件，跳过", dirPath)
			continue
		}

		// 随机选择一张图片
		selectedFile := imageFiles[rand.Intn(len(imageFiles))]

		// 读取并编码图片
		base64Img, err := c.loadAndEncodeImage(selectedFile)
		if err != nil {
			logx.Errorf("加载图片失败: %s, error: %v", selectedFile, err)
			continue
		}

		loadedImages = append(loadedImages, base64Img)
	}

	if len(loadedImages) == 0 {
		return nil, fmt.Errorf("没有成功加载任何图片")
	}

	return loadedImages, nil
}

// loadAndEncodeImage 加载并编码单张图片
func (c *ImageSelectCaptcha) loadAndEncodeImage(filePath string) (string, error) {
	// 打开图片文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 解码图片
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("解码图片失败: %w", err)
	}

	// 编码为 PNG 并转换为 Base64
	return c.encodeImageToBase64(img, 0)
}

// generatePlaceholderImages 生成占位图片（回退方案）
func (c *ImageSelectCaptcha) generatePlaceholderImages() ([]string, error) {
	var images []string

	for i := 0; i < c.config.ImageCount; i++ {
		// 创建占位图片
		img := image.NewRGBA(image.Rect(0, 0, 200, 150))

		// 白色背景
		draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)

		// 根据索引绘制不同的图形
		switch i {
		case 0:
			// 绘制类似公交车的图形（矩形 + 圆形轮子）
			// 车身
			for y := 50; y < 100; y++ {
				for x := 30; x < 170; x++ {
					img.Set(x, y, color.RGBA{200, 50, 50, 255}) // 红色车身
				}
			}
			// 车窗
			for y := 55; y < 75; y++ {
				for x := 40; x < 80; x++ {
					img.Set(x, y, color.RGBA{135, 206, 250, 255}) // 蓝色窗户
				}
				for x := 90; x < 160; x++ {
					img.Set(x, y, color.RGBA{135, 206, 250, 255})
				}
			}
			// 轮子
			drawCircle(img, 60, 100, 15, color.RGBA{50, 50, 50, 255})  // 左轮
			drawCircle(img, 140, 100, 15, color.RGBA{50, 50, 50, 255}) // 右轮

		case 1:
			// 绘制类似自行车的图形
			// 车架（简单的线条和形状）
			// 后轮
			drawCircle(img, 50, 90, 18, color.RGBA{50, 50, 50, 255})
			// 前轮
			drawCircle(img, 150, 90, 18, color.RGBA{50, 50, 50, 255})
			// 车架（菱形）
			for y := 70; y < 90; y++ {
				for x := 95; x < 105; x++ {
					img.Set(x, y, color.RGBA{100, 150, 255, 255})
				}
			}
			// 座位
			for y := 68; y < 72; y++ {
				for x := 80; x < 110; x++ {
					img.Set(x, y, color.RGBA{100, 150, 255, 255})
				}
			}
			// 把手
			for y := 55; y < 72; y++ {
				for x := 145; x < 155; x++ {
					img.Set(x, y, color.RGBA{100, 150, 255, 255})
				}
			}

		case 2:
			// 绘制红绿灯（竖直排列的圆形）
			// 外框
			for y := 30; y < 120; y++ {
				for x := 85; x < 115; x++ {
					img.Set(x, y, color.RGBA{80, 80, 80, 255})
				}
			}
			// 红灯（上）
			drawCircle(img, 100, 50, 12, color.RGBA{255, 0, 0, 255})
			// 黄灯（中）
			drawCircle(img, 100, 75, 12, color.RGBA{255, 255, 0, 255})
			// 绿灯（下）
			drawCircle(img, 100, 100, 12, color.RGBA{0, 255, 0, 255})

		case 3:
			// 绘制汽车的图形
			// 车身
			for y := 60; y < 110; y++ {
				for x := 40; x < 160; x++ {
					img.Set(x, y, color.RGBA{50, 100, 200, 255}) // 蓝色车身
				}
			}
			// 车顶
			for y := 45; y < 65; y++ {
				for x := 60; x < 140; x++ {
					img.Set(x, y, color.RGBA{50, 100, 200, 255})
				}
			}
			// 车窗
			for y := 50; y < 60; y++ {
				for x := 70; x < 100; x++ {
					img.Set(x, y, color.RGBA{135, 206, 250, 255})
				}
				for x := 110; x < 130; x++ {
					img.Set(x, y, color.RGBA{135, 206, 250, 255})
				}
			}
			// 轮子
			drawCircle(img, 70, 110, 12, color.RGBA{30, 30, 30, 255})
			drawCircle(img, 130, 110, 12, color.RGBA{30, 30, 30, 255})
		}

		// 添加图片标签（底部小字）
		face := basicfont.Face7x13
		labels := []string{"BUS", "BIKE", "LIGHT", "CAR"}

		drawer := &font.Drawer{
			Dst:  img,
			Src:  &image.Uniform{color.RGBA{0, 0, 0, 255}},
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(85), Y: fixed.I(135)},
		}
		drawer.DrawString(labels[i])

		base64Img, err := c.encodeImageToBase64(img, i)
		if err != nil {
			return nil, err
		}
		images = append(images, base64Img)
	}

	return images, nil
}

// drawCircle 辅助函数：绘制填充圆
func drawCircle(img *image.RGBA, centerX, centerY, radius int, c color.RGBA) {
	for y := centerY - radius; y <= centerY+radius; y++ {
		for x := centerX - radius; x <= centerX+radius; x++ {
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy <= radius*radius {
				if y >= 0 && y < 150 && x >= 0 && x < 200 {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// encodeImageToBase64 将图片编码为Base64
func (c *ImageSelectCaptcha) encodeImageToBase64(img image.Image, index int) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// generateQuestion 生成问题文本
func (c *ImageSelectCaptcha) generateQuestion() string {
	questions := []string{
		"请选择所有的公交车",
		"请选择所有的红绿灯",
		"请选择所有的自行车",
		"请选择所有的汽车",
	}

	rand.Seed(time.Now().UnixNano())
	return questions[rand.Intn(len(questions))]
}

// loadImageFromFile 从文件加载图片
func (c *ImageSelectCaptcha) loadImageFromFile(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// loadImagesFromDir 从目录加载图片
func (c *ImageSelectCaptcha) loadImagesFromDir(dir string, count int) ([]image.Image, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var images []image.Image
	for _, file := range files {
		if len(images) >= count {
			break
		}

		if !file.IsDir() {
			filePath := filepath.Join(dir, file.Name())
			img, err := c.loadImageFromFile(filePath)
			if err != nil {
				logx.Errorf("Failed to load image %s: %v", filePath, err)
				continue
			}
			images = append(images, img)
		}
	}

	if len(images) < count {
		return nil, fmt.Errorf("not enough images in directory, got %d, need %d", len(images), count)
	}

	// 随机打乱
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(images), func(i, j int) {
		images[i], images[j] = images[j], images[i]
	})

	return images, nil
}

// init 初始化
func init() {
	logx.Info("ImageSelect captcha initialized")
}
