// 全局状态
let currentCaptchaId = null;
let sliderStartTime = 0;
let isDragging = false;

// DOM 元素
const generateBtn = document.getElementById('generateBtn');
const refreshBtn = document.getElementById('refreshBtn');
const verifyBtn = document.getElementById('verifyBtn');
const captchaDisplay = document.getElementById('captchaDisplay');
const inputSection = document.getElementById('inputSection');
const resultSection = document.getElementById('resultSection');
const resultCard = document.getElementById('resultCard');
const resultIcon = document.getElementById('resultIcon');
const resultMessage = document.getElementById('resultMessage');

// 生成验证码
generateBtn.addEventListener('click', generateCaptcha);
refreshBtn.addEventListener('click', generateCaptcha);

// 验证验证码
verifyBtn.addEventListener('click', verifyCaptcha);

async function generateCaptcha() {
    try {
        const response = await fetch('/api/captcha/generate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        });

        const result = await response.json();

        if (result.code === 0) {
            currentCaptchaId = result.data.captchaId;
            displayCaptcha(result.data);
            inputSection.style.display = 'block';
            refreshBtn.style.display = 'inline-block';
            hideResult();
        } else {
            showError('生成验证码失败: ' + result.message);
        }
    } catch (error) {
        showError('网络错误: ' + error.message);
    }
}

function displayCaptcha(data) {
    captchaDisplay.innerHTML = `
        <div class="slide-container">
            <div class="slide-canvas">
                <img src="${data.backgroundImage}" alt="背景图" class="background-img">
                <img src="${data.tileImage}" alt="滑块" class="template-img" id="templateImg">
                <div class="slider-track">
                    <div class="slider-button" id="sliderBtn">
                        <span class="slider-icon">→</span>
                    </div>
                    <div class="slider-text">拖动滑块完成拼图</div>
                </div>
            </div>
        </div>
    `;

    // 延迟初始化，确保 DOM 已渲染
    setTimeout(() => {
        initSlider();
    }, 100);
}

function initSlider() {
    const sliderBtn = document.getElementById('sliderBtn');
    const templateImg = document.getElementById('templateImg');
    if (!sliderBtn || !templateImg) return;

    // 设置初始位置
    sliderBtn.style.left = '0px';
    templateImg.style.left = '0px';

    // 添加事件监听器
    sliderBtn.addEventListener('mousedown', startSlide);
    sliderBtn.addEventListener('touchstart', startSlide, { passive: false });
}

function startSlide(e) {
    e.preventDefault();

    isDragging = true;
    sliderStartTime = Date.now();

    const sliderBtn = e.target.closest('.slider-button');
    const templateImg = document.getElementById('templateImg');
    const startX = e.type === 'mousedown' ? e.clientX : e.touches[0].clientX;
    const initialLeft = parseInt(sliderBtn.style.left || 0);

    // 添加文档级事件监听
    document.addEventListener('mousemove', onSlide);
    document.addEventListener('mouseup', endSlide);
    document.addEventListener('touchmove', onSlide, { passive: false });
    document.addEventListener('touchend', endSlide);

    function onSlide(e) {
        if (!isDragging) return;

        const currentX = e.type === 'mousemove' ? e.clientX : e.touches[0].clientX;
        const diff = currentX - startX;

        // 限制滑动范围 (背景图宽度 - 滑块宽度)
        const maxSlide = 290; // 350px 背景宽度 - 60px 滑块宽度
        const newLeft = Math.max(0, Math.min(initialLeft + diff, maxSlide));

        sliderBtn.style.left = newLeft + 'px';

        // 同步移动模板图片
        if (templateImg) {
            templateImg.style.left = newLeft + 'px';
        }
    }

    function endSlide() {
        isDragging = false;

        document.removeEventListener('mousemove', onSlide);
        document.removeEventListener('mouseup', endSlide);
        document.removeEventListener('touchmove', onSlide);
        document.removeEventListener('touchend', endSlide);

        // 计算最终位置百分比
        const currentLeft = parseInt(sliderBtn.style.left || 0);
        const xPercent = Math.round((currentLeft / 290) * 100);
        const duration = Date.now() - sliderStartTime;

        // 存储滑动数据供验证使用
        window.sliderData = {
            x: xPercent,
            duration: duration
        };
    }
}

// 验证验证码
async function verifyCaptcha() {
    if (!currentCaptchaId) {
        showError('请先生成验证码');
        return;
    }

    if (!window.sliderData) {
        showError('请完成滑动验证');
        return;
    }

    try {
        const response = await fetch('/api/captcha/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                captchaId: currentCaptchaId,
                x: window.sliderData.x,
                duration: window.sliderData.duration
            })
        });

        const result = await response.json();
        showResult(result.data.valid);
    } catch (error) {
        showError('验证失败: ' + error.message);
    }
}

function showResult(success) {
    resultSection.style.display = 'block';
    resultCard.className = 'result-card ' + (success ? 'success' : 'error');
    resultIcon.textContent = success ? '✅' : '❌';
    resultMessage.textContent = success ? '验证成功！' : '验证失败，请重试！';

    if (success) {
        setTimeout(() => {
            resetUI();
        }, 2000);
    }
}

function showError(message) {
    resultSection.style.display = 'block';
    resultCard.className = 'result-card error';
    resultIcon.textContent = '⚠️';
    resultMessage.textContent = message;
}

function hideResult() {
    resultSection.style.display = 'none';
}

function resetUI() {
    currentCaptchaId = null;
    window.sliderData = null;
    captchaDisplay.innerHTML = '<p class="placeholder">点击下方按钮生成验证码</p>';
    inputSection.style.display = 'none';
    refreshBtn.style.display = 'none';
    hideResult();
}
