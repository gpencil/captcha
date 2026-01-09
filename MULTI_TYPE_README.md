# 多类型验证码服务使用指南

## 概述

验证码服务现在支持三种验证码方式：
1. **字符验证码**（character）- 传统字符图形验证码
2. **图片选择验证码**（image_select）- 点击图片中的特定物体
3. **滑动验证码**（slide）- 滑动拼图验证码

所有验证码类型使用同一个 API 接口，可通过配置文件轻松切换。

## 配置说明

### 配置文件（etc/dev/tts.yaml 或 etc/prod/tts.yaml）

```yaml
# 验证码配置
Captcha:
  Enabled: true  # 是否启用验证码
  Type: "character"  # 验证码类型选择
  # - character: 字符验证码（默认，简单易用）
  # - image_select: 图片选择验证码（更友好）
  # - slide: 滑动验证码（更安全）
```

## API 接口

所有验证码类型统一使用同一个生成和验证接口。

### 1. 生成验证码

**接口地址**：`POST /api/tts/v1/captcha/generate`

**请求参数**：
```json
{
  "captchaType": "character"  // 验证码类型
  // - "character": 字符验证码
  // - "image_select": 图片选择验证码
  // - "slide": 滑动验证码
}
```

#### 响应示例

**字符验证码响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "captchaId": "uuid-string",
    "captchaType": "character",
    "data": {
      "image": "data:image/png;base64,iVBORw0KG..."  // Base64图片
    },
    "expireTime": 1704356400
  }
}
```

**图片选择验证码响应**：
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

**滑动验证码响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "captchaId": "uuid-string",
    "captchaType": "slide",
    "data": {
      "backgroundImage": "data:image/png;base64,iVBORw0KG...",  // 背景图
      "templateImage": "data:image/png;base64,iVBORw0KG...",    // 滑块图
      "templateY": 50,       // 滑块Y轴位置
      "width": 350,          // 背景图宽度
      "height": 200          // 背景图高度
    },
    "expireTime": 1704356400
  }
}
```

### 2. 发送短信验证码（集成验证码验证）

**接口地址**：`POST /api/tts/v1/auth/code/send`

#### 字符验证码请求：
```json
{
  "phone": "13800138000",
  "areaCode": "+86",
  "captchaId": "uuid-string",    // 验证码ID
  "captchaCode": "ABCD"          // 用户输入的验证码
}
```

#### 图片选择验证码请求：
```json
{
  "phone": "13800138000",
  "areaCode": "+86",
  "captchaId": "uuid-string",    // 验证码ID
  "captchaAnswer": {              // 选中的图片索引
    "selectedIndexes": [0, 2]
  }
}
```

#### 滑动验证码请求：
```json
{
  "phone": "13800138000",
  "areaCode": "+86",
  "captchaId": "uuid-string",    // 验证码ID
  "captchaAnswer": {              // 滑动数据
    "x": 45,                      // 滑块X轴位置（百分比0-100）
    "track": [10, 20, 30, ...],  // 滑动轨迹
    "duration": 1500             // 滑动耗时（毫秒）
  }
}
```

## 验证码类型切换

### 开发环境

编辑 `etc/dev/tts.yaml`：

```yaml
Captcha:
  Enabled: true
  Type: "character"  # 可选: character, image_select, slide
```

### 生产环境

编辑 `etc/prod/tts.yaml`：

```yaml
Captcha:
  Enabled: true
  Type: "slide"  # 生产环境推荐使用滑动验证码
```

## 前端集成示例

### 通用验证码组件（支持切换）

```javascript
// 根据配置动态选择验证码类型
const CAPTCHA_TYPE = 'character'; // 可配置为 'image_select' 或 'slide'

// 生成验证码
async function generateCaptcha() {
  const response = await fetch('/api/tts/v1/captcha/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ captchaType: CAPTCHA_TYPE })
  });

  const result = await response.json();
  if (result.code === 0) {
    // 根据类型显示不同的验证码UI
    switch (CAPTCHA_TYPE) {
      case 'character':
        showCharacterCaptcha(result.data.data.image);
        break;
      case 'image_select':
        showImageSelectCaptcha(result.data.data);
        break;
      case 'slide':
        showSlideCaptcha(result.data.data);
        break;
    }
    // 保存验证码ID
    sessionStorage.setItem('captchaId', result.data.captchaId);
  }
}

// 字符验证码UI
function showCharacterCaptcha(imageBase64) {
  document.getElementById('captcha-container').innerHTML = `
    <img src="${imageBase64}" onclick="refreshCaptcha()" />
    <input type="text" id="captcha-code" placeholder="请输入验证码" />
  `;
}

// 图片选择验证码UI
function showImageSelectCaptcha(data) {
  let imagesHtml = data.images.map((img, index) => `
    <img src="${img}" onclick="selectImage(${index})" />
  `).join('');

  document.getElementById('captcha-container').innerHTML = `
    <p>${data.question}</p>
    <div class="image-grid">${imagesHtml}</div>
    <p>已选择: <span id="selected-count">0</span>/${data.selectCount}</p>
  `;
}

// 滑动验证码UI
function showSlideCaptcha(data) {
  document.getElementById('captcha-container').innerHTML = `
    <div class="slide-captcha">
      <img src="${data.backgroundImage}" class="background" />
      <img src="${data.templateImage}" class="template" />
      <div class="slider-track">
        <div class="slider-button" id="slider-btn"></div>
      </div>
    </div>
  `;

  initSlider();
}

// 发送短信验证码
async function sendSmsCode() {
  const phone = document.getElementById('phone').value;
  const captchaId = sessionStorage.getItem('captchaId');

  let body = {
    phone: phone,
    areaCode: '+86',
    captchaId: captchaId
  };

  // 根据验证码类型添加不同的答案
  switch (CAPTCHA_TYPE) {
    case 'character':
      body.captchaCode = document.getElementById('captcha-code').value;
      break;
    case 'image_select':
      body.captchaAnswer = {
        selectedIndexes: getSelectedIndexes()
      };
      break;
    case 'slide':
      body.captchaAnswer = {
        x: getSliderX(),
        track: getSliderTrack(),
        duration: getSliderDuration()
      };
      break;
  }

  const response = await fetch('/api/tts/v1/auth/code/send', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });

  const result = await response.json();
  if (result.code === 0) {
    alert('短信验证码已发送');
  } else {
    alert('发送失败：' + result.message);
  }
}
```

## 各验证码类型特点

### 1. 字符验证码（character）

**优点**：
- 实现简单，兼容性好
- 无需额外资源
- 用户体验传统熟悉

**缺点**：
- 可能被OCR识别
- 用户体验相对较差

**适用场景**：
- 内部管理系统
- 低风险场景
- 开发测试环境

**配置复杂度**：
```go
CharacterConfig{
    Width:       160,     // 图片宽度
    Height:      60,      // 图片高度
    Length:      4,       // 验证码长度
    ExpireTime:  5 * time.Minute,
    Complexity:  2,       // 1-简单, 2-中等, 3-复杂
}
```

### 2. 图片选择验证码（image_select）

**优点**：
- 用户友好，直观易用
- 比字符验证码更安全
- 移动端体验好

**缺点**：
- 需要准备图片资源
- 占用空间较大

**适用场景**：
- C端应用
- 移动端应用
- 中等风险场景

**配置参数**：
```go
ImageSelectConfig{
    ImageCount:  4,           // 选项图片数量
    SelectCount: 1,           // 需要选择的数量
    ExpireTime:  5 * time.Minute,
    Category:    "traffic",   // 图片类别
    ImageDir:    "/path/to/images",  // 图片目录
}
```

**图片资源准备**：
```
/images/
├── traffic/
│   ├── bus_1.jpg
│   ├── bus_2.jpg
│   ├── car_1.jpg
│   └── ...
├── animal/
└── food/
```

### 3. 滑动验证码（slide）

**优点**：
- 安全性最高
- 用户体验流畅
- 难以机器识别

**缺点**：
- 实现复杂
- 需要更多计算资源

**适用场景**：
- 重要操作
- 高风险场景
- 金融支付场景

**配置参数**：
```go
SlideConfig{
    Width:           350,  // 背景图宽度
    Height:          200,  // 背景图高度
    TemplateWidth:   60,   // 滑块宽度
    TemplateHeight:  60,   // 滑块高度
    ExpireTime:      5 * time.Minute,
    ImageDir:        "/path/to/backgrounds",
    TemplateDir:     "/path/to/templates",
}
```

## 切换验证码类型

### 方法1：修改配置文件（推荐）

```yaml
# 开发环境 - 使用字符验证码
Captcha:
  Enabled: true
  Type: "character"

# 测试环境 - 使用图片选择验证码
Captcha:
  Enabled: true
  Type: "image_select"

# 生产环境 - 使用滑动验证码
Captcha:
  Enabled: true
  Type: "slide"
```

### 方法2：运行时动态切换

```go
// 根据用户行为动态选择验证码类型
captchaType := "character"
if isHighRiskUser(phone) {
    captchaType = "slide"  // 高风险用户使用滑动验证码
}

resp, err := captchaService.Generate(ctx, captcha.CaptchaType(captchaType))
```

## 安全建议

1. **开发环境**：使用 `character`，方便测试
2. **测试环境**：使用 `image_select`，测试用户体验
3. **生产环境**：使用 `slide`，安全性最高
4. **动态切换**：根据风险等级动态选择验证码类型

## 性能优化

1. **CDN加速**：图片资源使用CDN分发
2. **缓存**：前端缓存验证码图片，减少重复请求
3. **限流**：限制验证码生成频率
4. **异步**：验证码验证异步处理，不阻塞主流程

## 故障排查

### 问题1：验证码验证失败

**检查项**：
1. 验证码类型配置是否正确
2. 验证码ID是否过期（5分钟）
3. 验证码是否已被使用（一次性）
4. 答案格式是否正确

### 问题2：图片验证码显示异常

**检查项**：
1. 图片资源目录是否正确
2. Redis连接是否正常
3. Base64编码是否正确

### 问题3：滑动验证码无法通过

**检查项**：
1. X坐标误差范围（默认±5像素）
2. 滑动轨迹是否合理（至少10个点）
3. 滑动时间是否合理（0.5-10秒）

## 技术支持

如有问题，请查看：
- go-common/captcha/README.md - 验证码服务详细文档
- tts-api/CAPTCHA_README.md - tts-api集成文档
