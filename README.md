# Captcha 多类型验证码服务

用于短信防刷的验证码服务，支持**三种验证码方式**，统一接口，灵活切换。

## ✨ 验证码类型

### 1. 字符验证码（character）
传统字符图形验证码，适合开发测试环境。

**特点**：
- ✅ 实现简单，兼容性好
- ✅ 无需额外资源
- ⚠️ 可能被OCR识别

**适用场景**：开发环境、内部系统

---

### 2. 图片选择验证码（image_select）
从多张图片中选择符合要求的图片，用户体验友好。

**特点**：
- ✅ 用户友好，直观易用
- ✅ 移动端体验好
- ✅ 比字符验证码更安全
- ⚠️ 需要准备图片资源

**适用场景**：C端应用、移动端

---

### 3. 滑动验证码（slide）
滑动拼图验证码，安全性最高。

**特点**：
- ✅ 安全性最高
- ✅ 难以机器识别
- ✅ 用户体验流畅
- ⚠️ 实现较复杂

**适用场景**：金融支付、高风险场景

---

## 功能特性

- ✅ **三种验证码类型**：character / image_select / slide
- ✅ **统一接口**：一个API支持所有验证码类型
- ✅ **灵活切换**：配置文件一键切换验证码类型
- ✅ **动态选择**：运行时可动态选择验证码类型
- ✅ **Redis 存储**：支持分布式部署
- ✅ **内存存储**：支持单机测试
- ✅ **一次一密**：验证成功后自动删除
- ✅ **过期机制**：5分钟自动过期
- ✅ **无缝集成**：与短信服务无缝集成

## 安装依赖

```bash
go get golang.org/x/image/font
go get github.com/google/uuid
go get github.com/redis/go-redis/v9
```

## 快速开始

### 1. 初始化服务

```go
import (
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/gpencil/captcha"
)

// 创建 Redis 客户端
redisClient := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// 创建验证码存储
store := captcha.NewRedisStore(redisClient, "captcha:")

// 创建验证码服务（支持所有三种类型）
service := captcha.NewService(
    store,
    captcha.CharacterConfig{      // 字符验证码配置
        Width:       160,
        Height:      60,
        Length:      4,
        ExpireTime:  5 * time.Minute,
        Complexity:  2,
    },
    captcha.ImageSelectConfig{    // 图片选择验证码配置
        ImageCount:  4,
        SelectCount: 1,
        ExpireTime:  5 * time.Minute,
        Category:    "traffic",
        ImageDir:    "/path/to/images",
    },
    captcha.SlideConfig{           // 滑动验证码配置
        Width:           350,
        Height:          200,
        TemplateWidth:   60,
        TemplateHeight:  60,
        ExpireTime:      5 * time.Minute,
        ImageDir:        "/path/to/backgrounds",
        TemplateDir:     "/path/to/templates",
    },
)
```

### 2. 生成验证码（统一接口）

```go
// 生成字符验证码
resp, err := service.Generate(ctx, captcha.CaptchaTypeCharacter)

// 生成图片选择验证码
resp, err := service.Generate(ctx, captcha.CaptchaTypeImageSelect)

// 生成滑动验证码
resp, err := service.Generate(ctx, captcha.CaptchaType("slide"))
```

### 3. 验证验证码

```go
// 字符验证码验证
req := &captcha.VerifyRequest{
    CaptchaID:   "captcha-id",
    CaptchaType: captcha.CaptchaTypeCharacter,
    Answer: captcha.CharacterAnswer{
        Code: "ABCD",
    },
}

// 图片选择验证码验证
req := &captcha.VerifyRequest{
    CaptchaID:   "captcha-id",
    CaptchaType: captcha.CaptchaTypeImageSelect,
    Answer: captcha.ImageSelectAnswer{
        SelectedIndexes: []int{0, 2},
    },
}

// 滑动验证码验证
req := &captcha.VerifyRequest{
    CaptchaID:   "captcha-id",
    CaptchaType: captcha.CaptchaType("slide"),
    Answer: captcha.SlideAnswer{
        X:        45,          // 滑块位置（百分比）
        Track:    []int{...},  // 滑动轨迹
        Duration: 1500,        // 滑动耗时（毫秒）
    },
}

// 验证（一次性，验证后自动删除）
valid, err := service.VerifyAndDelete(ctx, req)
```

## API 接口

### 生成验证码

**接口**：`POST /api/tts/v1/captcha/generate`

**请求参数**：
```json
{
  "captchaType": "character"  // 验证码类型
  // "character" - 字符验证码
  // "image_select" - 图片选择验证码
  // "slide" - 滑动验证码
}
```

**响应示例（字符验证码）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "captchaId": "uuid-string",
    "captchaType": "character",
    "data": {
      "image": "data:image/png;base64,iVBORw0KG..."
    },
    "expireTime": 1704356400
  }
}
```

**响应示例（图片选择验证码）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "captchaId": "uuid-string",
    "captchaType": "image_select",
    "data": {
      "question": "请选择所有的公交车",
      "targetType": "bus",
      "images": [
        "data:image/png;base64,iVBORw0KG...",
        "data:image/png;base64,iVBORw0KG...",
        "data:image/png;base64,iVBORw0KG...",
        "data:image/png;base64,iVBORw0KG..."
      ],
      "selectCount": 1
    },
    "expireTime": 1704356400
  }
}
```

**响应示例（滑动验证码）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "captchaId": "uuid-string",
    "captchaType": "slide",
    "data": {
      "backgroundImage": "data:image/png;base64,iVBORw0KG...",
      "templateImage": "data:image/png;base64,iVBORw0KG...",
      "templateY": 50,
      "width": 350,
      "height": 200
    },
    "expireTime": 1704356400
  }
}
```

## 配置说明

### CharacterConfig（字符验证码配置）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Width | int | 160 | 图片宽度（像素） |
| Height | int | 60 | 图片高度（像素） |
| Length | int | 4 | 验证码长度 |
| ExpireTime | Duration | 5分钟 | 过期时间 |
| Complexity | int | 2 | 复杂度（1-简单，2-中等，3-复杂）|

**复杂度说明**：
- Level 1（简单）：只包含数字 `0-9`
- Level 2（中等）：数字+大写字母，去掉易混淆字符
- Level 3（复杂）：数字+大小写字母，去掉易混淆字符

---

### ImageSelectConfig（图片选择验证码配置）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| ImageCount | int | 4 | 选项图片数量 |
| SelectCount | int | 1 | 需要选择的数量 |
| ExpireTime | Duration | 5分钟 | 过期时间 |
| Category | string | - | 图片类别（traffic/animal/food）|
| ImageDir | string | - | 图片文件目录路径 |

---

### SlideConfig（滑动验证码配置）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Width | int | 350 | 背景图宽度（像素）|
| Height | int | 200 | 背景图高度（像素）|
| TemplateWidth | int | 60 | 滑块模板宽度（像素）|
| TemplateHeight | int | 60 | 滑块模板高度（像素）|
| ExpireTime | Duration | 5分钟 | 过期时间 |
| ImageDir | string | - | 背景图片目录路径 |
| TemplateDir | string | - | 滑块模板目录路径 |

## 验证码类型选择

### 推荐使用场景

| 场景 | 推荐类型 | 理由 |
|------|----------|------|
| 开发环境 | character | 简单快速，方便测试 |
| 测试环境 | image_select | 体验友好，便于测试 |
| 生产环境 | slide | 安全性最高 |
| C端应用 | image_select | 用户友好，移动端体验好 |
| 金融支付 | slide | 高安全性要求 |
| 内部系统 | character | 简单易用 |

### 切换方式

**方式1：配置文件切换**
```yaml
# etc/dev/tts.yaml
Captcha:
  Enabled: true
  Type: "character"  # 可选: character, image_select, slide
```

**方式2：运行时动态选择**
```go
// 根据场景选择验证码类型
captchaType := captcha.CaptchaTypeCharacter

if isHighRiskUser() {
    captchaType = "slide"  // 高风险用户使用滑动验证码
} else if isMobileClient() {
    captchaType = captcha.CaptchaTypeImageSelect  // 移动端使用图片选择
}

resp, err := service.Generate(ctx, captchaType)
```

## 使用示例

### 示例1：发送短信前验证验证码

```go
func SendSMSWithCaptcha(ctx context.Context, phone, captchaID, captchaCode string) error {
    // 1. 验证验证码
    req := &captcha.VerifyRequest{
        CaptchaID:   captchaID,
        CaptchaType: captcha.CaptchaTypeCharacter,
        Answer: captcha.CharacterAnswer{
            Code: captchaCode,
        },
    }

    valid, err := captchaService.VerifyAndDelete(ctx, req)
    if err != nil || !valid {
        return errors.New("验证码错误")
    }

    // 2. 发送短信
    _, err = smsClient.Send(ctx, &sms.SendRequest{
        Phone:   phone,
        BizID:   "login",
        Template: "SMS_TEMPLATE",
        Params: map[string]string{
            "code": "123456",
        },
    })

    return err
}
```

### 示例2：根据用户等级选择验证码

```go
func GenerateCaptchaForUser(ctx context.Context, userLevel int) (*captcha.CaptchaResponse, error) {
    var captchaType captcha.CaptchaType

    // 根据用户等级选择验证码类型
    switch userLevel {
    case 1: // 普通用户
        captchaType = captcha.CaptchaTypeCharacter
    case 2: // VIP用户
        captchaType = captcha.CaptchaTypeImageSelect
    case 3: // 高风险用户
        captchaType = captcha.CaptchaType("slide")
    default:
        captchaType = captcha.CaptchaTypeCharacter
    }

    return captchaService.Generate(ctx, captchaType)
}
```

## 防刷策略

1. **验证码有效期**：默认 5 分钟，超时需重新获取
2. **一次性使用**：验证成功后自动删除，防止重复使用
3. **多种验证码**：根据场景选择合适的验证码类型
4. **可配置难度**：根据需求调整验证码难度
5. **分布式支持**：Redis 存储，支持多实例部署

## 最佳实践

### 1. 前端集成
- 在短信发送按钮前增加验证码输入框
- 点击"获取验证码"前先要求用户完成图形验证码
- 只有图形验证码通过后才调用发送短信接口

### 2. 限流策略
- 结合 IP 限流：每个 IP 每分钟最多获取 N 次验证码
- 结合手机号限流：每个手机号每天最多发送 N 条短信
- 使用现有的 `rateLimiter` 包

### 3. 安全性
- 验证码使用后立即删除
- 使用 HTTPS 传输
- 不要在前端暴露真实的短信验证码
- 生产环境推荐使用滑动验证码

### 4. 性能优化
- 图片资源使用 CDN 加速
- 前端缓存验证码图片，减少重复请求
- 限制验证码生成频率
- 验证码验证异步处理

## 图片资源准备

### 图片选择验证码

准备图片资源目录结构：
```
/images/
├── traffic/
│   ├── bus_1.jpg
│   ├── bus_2.jpg
│   ├── car_1.jpg
│   └── ...
├── animal/
│   ├── cat_1.jpg
│   ├── dog_1.jpg
│   └── ...
└── food/
    ├── apple_1.jpg
    ├── banana_1.jpg
    └── ...
```

### 滑动验证码

准备背景图和滑块模板：
```
/slide/
├── backgrounds/
│   ├── bg_1.jpg
│   ├── bg_2.jpg
│   └── ...
└── templates/
    ├── template_1.png
    ├── template_2.png
    └── ...
```

## 故障排查

### 问题1：验证码验证失败

**可能原因**：
- 验证码ID过期（5分钟）
- 验证码已被使用（一次性）
- 答案格式错误

**解决方案**：
- 检查验证码生成时间
- 重新生成验证码
- 检查答案格式

### 问题2：图片验证码显示异常

**可能原因**：
- 图片目录配置错误
- 图片资源不存在
- Redis连接异常

**解决方案**：
- 检查 ImageDir 配置
- 确认图片资源存在
- 检查 Redis 连接

### 问题3：滑动验证码无法通过

**可能原因**：
- X坐标误差过大
- 滑动轨迹不合理
- 滑动时间异常

**解决方案**：
- 调整误差范围（默认±5像素）
- 确保滑动轨迹记录完整
- 滑动时间应在 0.5-10 秒之间

## 相关文档

- [多类型验证码使用指南](MULTI_TYPE_README.md) - 详细使用说明
- [tts-api集成文档](../tts-api/CAPTCHA_README.md) - tts-api集成示例

## 技术支持

如有问题，请提交 Issue 或联系开发团队。
