//go:build ignore
// +build ignore

package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 创建目录结构
	dirs := []string{
		"images/traffic",
		"images/backgrounds",
		"images/templates",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(err)
		}
	}

	// 生成交通工具图片（4张）
	generateTrafficImages()

	// 生成背景图片（3张）
	generateBackgroundImages()

	// 生成滑块模板（3张）
	generateTemplateImages()

	log.Println("测试图片生成完成！")
}

func generateTrafficImages() {
	colors := []color.RGBA{
		{255, 0, 0, 255},   // 红色
		{0, 255, 0, 255},   // 绿色
		{0, 0, 255, 255},   // 蓝色
		{255, 255, 0, 255}, // 黄色
	}

	labels := []string{"公交车", "汽车", "自行车", "摩托车"}

	for i := 0; i < 4; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 200, 150))

		// 填充背景
		draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{240, 240, 240, 255}}, image.Point{}, draw.Src)

		// 绘制彩色矩形代表交通工具
		rectColor := colors[i]
		draw.Draw(img, image.Rect(20, 40, 180, 110), &image.Uniform{rectColor}, image.Point{}, draw.Src)

		// 保存图片
		f, _ := os.Create(filepath.Join("images/traffic", trafficFileName(i)))
		defer f.Close()
		jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
		log.Printf("生成: %s\n", trafficFileName(i))
	}
}

func generateBackgroundImages() {
	for i := 0; i < 3; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 350, 200))

		// 创建渐变背景
		for y := 0; y < 200; y++ {
			for x := 0; x < 350; x++ {
				r := uint8((x * 255) / 350)
				g := uint8((y * 255) / 200)
				b := uint8(150 + i*30)
				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}

		// 添加一些随机图形
		for j := 0; j < 5; j++ {
			x := (i*50 + j*70) % 350
			y := (i*30 + j*40) % 200
			drawCircle(img, x, y, 20, color.RGBA{255, 255, 255, 200})
		}

		// 保存图片
		f, _ := os.Create(filepath.Join("images/backgrounds", backgroundFileName(i)))
		defer f.Close()
		jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
		log.Printf("生成: %s\n", backgroundFileName(i))
	}
}

func generateTemplateImages() {
	for i := 0; i < 3; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 60, 60))

		// 创建透明背景的滑块形状
		for y := 0; y < 60; y++ {
			for x := 0; x < 60; x++ {
				img.Set(x, y, color.RGBA{0, 0, 0, 0}) // 透明
			}
		}

		// 绘制拼图形状（使用白色）
		templateColor := color.RGBA{255, 255, 255, 255}
		for y := 5; y < 55; y++ {
			for x := 5; x < 55; x++ {
				// 简单的拼图形状
				if x >= 5 && x <= 55 && y >= 5 && y <= 55 {
					if (x <= 15 || x >= 45 || y <= 15 || y >= 45) ||
						(x >= 25 && x <= 35 && y >= 25 && y <= 35) {
						img.Set(x, y, templateColor)
					}
				}
			}
		}

		// 保存为 PNG（支持透明）
		f, _ := os.Create(filepath.Join("images/templates", templateFileName(i)))
		defer f.Close()
		png.Encode(f, img)
		log.Printf("生成: %s\n", templateFileName(i))
	}
}

func drawCircle(img *image.RGBA, x, y, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				ix, iy := x+dx, y+dy
				if ix >= 0 && ix < 350 && iy >= 0 && iy < 200 {
					img.Set(ix, iy, c)
				}
			}
		}
	}
}

func trafficFileName(i int) string {
	names := []string{"bus_1.jpg", "car_1.jpg", "bus_2.jpg", "bike_1.jpg"}
	return names[i]
}

func backgroundFileName(i int) string {
	return "bg_" + string(rune('1'+i)) + ".jpg"
}

func templateFileName(i int) string {
	return "template_" + string(rune('1'+i)) + ".png"
}
