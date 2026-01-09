# 验证码 Web 测试平台

这是一个用于测试 go-common/captcha 包三种验证码类型的 Web 应用。

## 功能特性

- ✅ **字符验证码** - 传统的字母数字图形验证码
- ✅ **图片选择验证码** - 从多张图片中选择符合要求的图片
- ✅ **滑动验证码** - 拖动滑块完成拼图验证
- ✅ 实时生成和验证
- ✅ 响应式设计，支持移动端
- ✅ 美观的用户界面

## 快速开始

### 1. 准备工作

确保你已经：
- 安装了 Go (1.16+)
- 安装了 Redis 并启动服务
- 克隆了 go-common 项目

### 2. 创建测试图片资源（可选）

如果需要测试图片选择和滑动验证码，需要准备测试图片：

```bash
# 创建图片目录
mkdir -p images/traffic
mkdir -p images/backgrounds
mkdir -p images/templates

# 放置测试图片（可以先用任何图片测试）
# - traffic/: 放置交通工具图片
# - backgrounds/: 放置背景图片（350x200）
# - templates/: 放置滑块模板图片（60x60，PNG格式）
```

### 3. 启动服务

```bash
cd go-common/captcha-web
go run main.go handlers.go
```

服务将在 `http://localhost:8080` 启动。

### 4. 访问测试

在浏览器中打开：`http://localhost:8080`

## 使用说明

### 字符验证码

1. 点击顶部"字符验证码"按钮
2. 点击"生成验证码"按钮
3. 查看图片中的字符
4. 在输入框中输入字符
5. 点击"验证"按钮

### 图片选择验证码

1. 点击顶部"图片选择"按钮
2. 点击"生成验证码"按钮
3. 根据问题点击选择相应的图片（可以多选）
4. 点击"验证"按钮

### 滑动验证码

1. 点击顶部"滑动验证"按钮
2. 点击"生成验证码"按钮
3. 拖动底部滑块到合适位置
4. 点击"验证"按钮

## 项目结构

```
captcha-web/
├── main.go           # 主程序，启动 HTTP 服务器
├── handlers.go       # HTTP 请求处理逻辑
├── static/
│   ├── index.html    # 前端页面
│   ├── app.js        # 前端交互逻辑
│   └── style.css     # 样式文件
├── images/           # 图片资源（可选）
│   ├── traffic/      # 图片选择验证码图片
│   ├── backgrounds/  # 滑动验证码背景图
│   └── templates/    # 滑动验证码滑块模板
└── README.md         # 说明文档
```

## API 接口

### 生成验证码

**请求**
```
POST /api/captcha/generate
Content-Type: application/json

{
  "captchaType": "character"  // 或 "image_select", "slide"
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "captchaId": "uuid",
    "captchaType": "character",
    "data": {
      "image": "data:image/png;base64,..."
    },
    "expireTime": 1234567890
  }
}
```

### 验证验证码

**请求**
```
POST /api/captcha/verify
Content-Type: application/json

{
  "captchaId": "uuid",
  "captchaType": "character",
  "captchaCode": "ABCD"  // 字符验证码
  // 或
  "captchaAnswer": {
    "selectedIndexes": [0, 2]  // 图片选择验证码
  }
  // 或
  "captchaAnswer": {
    "x": 45,
    "track": [10, 20, 30],
    "duration": 1500  // 滑动验证码
  }
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "valid": true
}
```

## 配置说明

### Redis 配置

在 `main.go` 中修改 Redis 连接信息：

```go
redisClient := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",  // Redis 地址
    Password: "",                  // Redis 密码
    DB:       0,                   // 数据库编号
})
```

### 验证码配置

在 `main.go` 中可以调整各种验证码的参数：

```go
// 字符验证码配置
captcha.CharacterConfig{
    Width:       160,     // 图片宽度
    Height:      60,      // 图片高度
    Length:      4,       // 验证码长度
    ExpireTime:  5 * 60,  // 过期时间（秒）
    Complexity:  2,       // 复杂度 1-3
}

// 图片选择验证码配置
captcha.ImageSelectConfig{
    ImageCount:  4,           // 选项数量
    SelectCount: 1,           // 需要选择的数量
    ExpireTime:  5 * 60,
    Category:    "traffic",
    ImageDir:    "./images/traffic",
}

// 滑动验证码配置
captcha.SlideConfig{
    Width:           350,   // 背景图宽度
    Height:          200,   // 背景图高度
    TemplateWidth:   60,    // 滑块宽度
    TemplateHeight:  60,    // 滑块高度
    ExpireTime:      5 * 60,
    ImageDir:        "./images/backgrounds",
    TemplateDir:     "./images/templates",
}
```

### 服务端口配置

在 `main.go` 中修改监听端口：

```go
addr := ":8080"  // 修改为其他端口
```

## 技术栈

- **后端**: Go + go-common/captcha
- **前端**: 原生 HTML/CSS/JavaScript
- **数据库**: Redis（用于存储验证码）

## 注意事项

1. **Redis 连接**: 确保 Redis 服务已启动并可连接
2. **图片资源**: 图片选择和滑动验证码需要准备相应的图片资源
3. **验证码过期**: 默认 5 分钟过期，需要在此时间内完成验证
4. **一次性使用**: 验证码验证成功后会自动删除，不能重复使用

## 常见问题

### Q: 生成的验证码显示为空白？

A: 检查图片目录配置是否正确，确保图片文件存在且可读。

### Q: 验证总是失败？

A: 检查 Redis 连接是否正常，验证码是否已过期。

### Q: 滑动验证码无法拖动？

A: 检查浏览器是否支持 JavaScript，尝试使用现代浏览器（Chrome、Firefox、Safari）。

## 开发建议

1. 测试时可以先使用字符验证码，不需要准备图片资源
2. 可以调整验证码配置参数以适应不同场景
3. 建议在开发环境使用较低的复杂度设置
4. 生产环境建议使用 HTTPS 协议

## 许可证

本测试平台仅供开发和测试使用。
