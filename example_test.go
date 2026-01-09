package captcha_test

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/gpencil/captcha"
)

// Example_character 字符验证码使用示例
func Example_character() {
	ctx := context.Background()

	// 1. 创建 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 2. 创建验证码存储
	store := captcha.NewRedisStore(redisClient, "captcha:")

	// 3. 创建验证码服务（支持所有三种类型）
	service := captcha.NewService(
		store,
		captcha.CharacterConfig{
			Width:      160,
			Height:     60,
			Length:     4,
			ExpireTime: 5 * time.Minute,
			Complexity: 2,
		},
		captcha.ImageSelectConfig{
			ImageCount:  4,
			SelectCount: 1,
			ExpireTime:  5 * time.Minute,
			Category:    "traffic",
			ImageDir:    "/path/to/images",
		},
		captcha.SlideConfig{
			Width:          350,
			Height:         200,
			TemplateWidth:  60,
			TemplateHeight: 60,
			ExpireTime:     5 * time.Minute,
			ImageDir:       "/path/to/backgrounds",
			TemplateDir:    "/path/to/templates",
		},
	)

	// 4. 生成字符验证码
	resp, err := service.Generate(ctx, captcha.CaptchaTypeCharacter)
	if err != nil {
		panic(err)
	}

	_ = resp // resp.CaptchaID, resp.Data, resp.ExpireTime

	// 5. 验证字符验证码
	verifyReq := &captcha.VerifyRequest{
		CaptchaID:   resp.CaptchaID,
		CaptchaType: captcha.CaptchaTypeCharacter,
		Answer: captcha.CharacterAnswer{
			Code: "ABCD", // 用户输入
		},
	}

	valid, err := service.VerifyAndDelete(ctx, verifyReq)
	if err != nil {
		panic(err)
	}

	_ = valid // true or false
}

// Example_imageSelect 图片选择验证码使用示例
func Example_imageSelect() {
	ctx := context.Background()

	// 创建验证码服务（参考 ExampleCharacterCaptcha）
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	store := captcha.NewRedisStore(redisClient, "captcha:")
	service := captcha.NewService(
		store,
		captcha.CharacterConfig{},
		captcha.ImageSelectConfig{
			ImageCount:  4,
			SelectCount: 1,
			ExpireTime:  5 * time.Minute,
			Category:    "traffic",
			ImageDir:    "/path/to/images",
		},
		captcha.SlideConfig{},
	)

	// 1. 生成图片选择验证码
	resp, err := service.Generate(ctx, captcha.CaptchaTypeImageSelect)
	if err != nil {
		panic(err)
	}

	_ = resp.CaptchaID  // 验证码ID
	_ = resp.ExpireTime // 过期时间
	// resp.Data 包含:
	//   - Question: "请选择所有的公交车"
	//   - TargetType: "bus"
	//   - Images: []string{base64Images}
	//   - SelectCount: 1

	// 2. 验证图片选择验证码
	verifyReq := &captcha.VerifyRequest{
		CaptchaID:   resp.CaptchaID,
		CaptchaType: captcha.CaptchaTypeImageSelect,
		Answer: captcha.ImageSelectAnswer{
			SelectedIndexes: []int{0, 2}, // 用户选择的图片索引
		},
	}

	valid, err := service.VerifyAndDelete(ctx, verifyReq)
	if err != nil {
		panic(err)
	}

	_ = valid // true or false
}

// Example_slide 滑动验证码使用示例
func Example_slide() {
	ctx := context.Background()

	// 创建验证码服务（参考 ExampleCharacterCaptcha）
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	store := captcha.NewRedisStore(redisClient, "captcha:")
	service := captcha.NewService(
		store,
		captcha.CharacterConfig{},
		captcha.ImageSelectConfig{},
		captcha.SlideConfig{
			Width:          350,
			Height:         200,
			TemplateWidth:  60,
			TemplateHeight: 60,
			ExpireTime:     5 * time.Minute,
			ImageDir:       "/path/to/backgrounds",
			TemplateDir:    "/path/to/templates",
		},
	)

	// 1. 生成滑动验证码
	resp, err := service.Generate(ctx, captcha.CaptchaType("slide"))
	if err != nil {
		panic(err)
	}

	_ = resp.CaptchaID  // 验证码ID
	_ = resp.ExpireTime // 过期时间
	// resp.Data 包含:
	//   - BackgroundImage: base64背景图
	//   - TemplateImage: base64滑块图
	//   - TemplateY: 滑块Y轴位置
	//   - Width, Height: 图片尺寸

	// 2. 验证滑动验证码
	verifyReq := &captcha.VerifyRequest{
		CaptchaID:   resp.CaptchaID,
		CaptchaType: captcha.CaptchaType("slide"),
		Answer: captcha.SlideAnswer{
			X:        45,                        // 滑块X轴位置（百分比）
			Track:    []int{10, 20, 30, 40, 50}, // 滑动轨迹
			Duration: 1500,                      // 滑动耗时（毫秒）
		},
	}

	valid, err := service.VerifyAndDelete(ctx, verifyReq)
	if err != nil {
		panic(err)
	}

	_ = valid // true or false
}
