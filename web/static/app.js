// 全局状态
let currentCaptchaType = 'character';
let currentCaptchaId = null;
let selectedIndexes = [];

// DOM 元素
const generateBtn = document.getElementById('generateBtn');
const refreshBtn = document.getElementById('refreshBtn');
const verifyBtn = document.getElementById('verifyBtn');
const captchaDisplay = document.getElementById('captchaDisplay');
const inputSection = document.getElementById('inputSection');
const characterInput = document.getElementById('characterInput');
const imageSelectInput = document.getElementById('imageSelectInput');
const slideInput = document.getElementById('slideInput');
const resultSection = document.getElementById('resultSection');
const resultCard = document.getElementById('resultCard');
const resultIcon = document.getElementById('resultIcon');
const resultMessage = document.getElementById('resultMessage');

// 验证码类型选择器
document.querySelectorAll('.type-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.type-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentCaptchaType = btn.dataset.type;
        resetUI();
    });
});

// 生成验证码
generateBtn.addEventListener('click', generateCaptcha);
refreshBtn.addEventListener('click', generateCaptcha);

// 验证验证码
verifyBtn.addEventListener('click', verifyCaptcha);

async function generateCaptcha() {
    try {
        const response = await fetch('/api/captcha/generate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ captchaType: currentCaptchaType })
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
    switch (currentCaptchaType) {
        case 'character':
            displayCharacterCaptcha(data.data);
            break;
        case 'image_select':
            displayImageSelectCaptcha(data.data);
            break;
        case 'slide':
            displaySlideCaptcha(data.data);
            break;
    }
}

// 显示字符验证码
function displayCharacterCaptcha(data) {
    characterInput.style.display = 'block';
    imageSelectInput.style.display = 'none';
    slideInput.style.display = 'none';

    captchaDisplay.innerHTML = `
        <div class="captcha-canvas">
            <img src="${data.image}" alt="验证码" onclick="generateCaptcha()" style="cursor: pointer;">
            <p class="hint">点击图片刷新</p>
        </div>
    `;
}

// 显示图片选择验证码
function displayImageSelectCaptcha(data) {
    characterInput.style.display = 'none';
    imageSelectInput.style.display = 'block';
    slideInput.style.display = 'none';
    selectedIndexes = [];
    updateSelectedCount();

    const imagesHtml = data.images.map((img, index) => `
        <div class="image-item ${selectedIndexes.includes(index) ? 'selected' : ''}"
             onclick="toggleImageSelection(${index})"
             data-index="${index}">
            <img src="${img}" alt="图片 ${index + 1}">
        </div>
    `).join('');

    captchaDisplay.innerHTML = `
        <div class="image-select-container">
            <div class="question">${data.question}</div>
            <div class="image-grid">
                ${imagesHtml}
            </div>
            <p class="select-count">需要选择: ${data.selectCount} 张</p>
        </div>
    `;
}

// 切换图片选择状态
window.toggleImageSelection = function(index) {
    const idx = selectedIndexes.indexOf(index);
    if (idx > -1) {
        selectedIndexes.splice(idx, 1);
    } else {
        selectedIndexes.push(index);
    }

    // 更新UI
    const imageItems = document.querySelectorAll('.image-item');
    imageItems.forEach(item => {
        const itemIndex = parseInt(item.dataset.index);
        if (selectedIndexes.includes(itemIndex)) {
            item.classList.add('selected');
        } else {
            item.classList.remove('selected');
        }
    });

    updateSelectedCount();
};

function updateSelectedCount() {
    document.getElementById('selectedCount').textContent = selectedIndexes.length;
}

// 显示滑动验证码
function displaySlideCaptcha(data) {
    characterInput.style.display = 'none';
    imageSelectInput.style.display = 'none';
    slideInput.style.display = 'block';

    // 清空之前的滑块数据
    window.sliderData = null;

    captchaDisplay.innerHTML = `
        <div class="slide-container">
            <div class="slide-canvas">
                <img src="${data.backgroundImage}" alt="背景图" class="background-img">
                <div class="slider-track">
                    <div class="slider-button" id="sliderBtn">→</div>
                </div>
            </div>
            <p class="hint">拖动滑块完成拼图</p>
        </div>
    `;

    // 延迟初始化，确保 DOM 已渲染
    setTimeout(() => {
        initSlider();
    }, 100);
}

let sliderStartX = 0;
let sliderTrack = [];
let sliderStartTime = 0;
let isDragging = false;

function initSlider() {
    const sliderBtn = document.getElementById('sliderBtn');
    if (!sliderBtn) return;

    console.log('初始化滑块');

    // 重置状态
    sliderTrack = [];
    sliderStartTime = 0;
    isDragging = false;

    // 设置初始位置
    sliderBtn.style.left = '0px';

    // 移除旧的事件监听器
    sliderBtn.removeEventListener('mousedown', startSlide);
    sliderBtn.removeEventListener('touchstart', startSlide);

    // 添加新的事件监听器
    sliderBtn.addEventListener('mousedown', startSlide);
    sliderBtn.addEventListener('touchstart', startSlide, { passive: false });
}

function startSlide(e) {
    e.preventDefault();
    console.log('开始拖动滑块');

    isDragging = true;
    sliderStartTime = Date.now();
    sliderTrack = [];

    const sliderBtn = e.target;
    const startX = e.type === 'mousedown' ? e.clientX : e.touches[0].clientX;
    const initialLeft = parseInt(sliderBtn.style.left || 0);

    console.log('起始位置:', startX, '初始 left:', initialLeft);

    // 添加文档级事件监听
    document.addEventListener('mousemove', onSlide);
    document.addEventListener('mouseup', endSlide);
    document.addEventListener('touchmove', onSlide, { passive: false });
    document.addEventListener('touchend', endSlide);

    function onSlide(e) {
        if (!isDragging) return;

        const currentX = e.type === 'mousemove' ? e.clientX : e.touches[0].clientX;
        const diff = currentX - startX;

        // 记录轨迹
        if (sliderTrack.length === 0 || Math.abs(currentX - sliderTrack[sliderTrack.length - 1]) > 2) {
            sliderTrack.push(currentX);
        }

        // 限制滑动范围
        const maxSlide = 280;
        const newLeft = Math.max(0, Math.min(initialLeft + diff, maxSlide));

        sliderBtn.style.left = newLeft + 'px';

        console.log('当前位置:', newLeft);
    }

    function endSlide() {
        console.log('结束拖动');
        isDragging = false;

        document.removeEventListener('mousemove', onSlide);
        document.removeEventListener('mouseup', endSlide);
        document.removeEventListener('touchmove', onSlide);
        document.removeEventListener('touchend', endSlide);

        // 计算最终位置百分比
        const currentLeft = parseInt(sliderBtn.style.left || 0);
        const xPercent = Math.round((currentLeft / 280) * 100);

        console.log('最终位置:', currentLeft, '百分比:', xPercent, '轨迹点数:', sliderTrack.length, '耗时:', Date.now() - sliderStartTime);

        // 存储滑动数据供验证使用
        window.sliderData = {
            x: xPercent,
            track: sliderTrack,
            duration: Date.now() - sliderStartTime
        };
    }
}

// 验证验证码
async function verifyCaptcha() {
    if (!currentCaptchaId) {
        showError('请先生成验证码');
        return;
    }

    let requestBody = {
        captchaId: currentCaptchaId,
        captchaType: currentCaptchaType
    };

    // 根据类型添加答案
    switch (currentCaptchaType) {
        case 'character':
            const code = document.getElementById('captchaCode').value.trim();
            if (!code) {
                showError('请输入验证码');
                return;
            }
            requestBody.captchaCode = code;
            break;

        case 'image_select':
            if (selectedIndexes.length === 0) {
                showError('请选择图片');
                return;
            }
            requestBody.captchaAnswer = {
                selectedIndexes: selectedIndexes
            };
            break;

        case 'slide':
            if (!window.sliderData) {
                showError('请完成滑动验证');
                return;
            }
            requestBody.captchaAnswer = window.sliderData;
            break;
    }

    try {
        const response = await fetch('/api/captcha/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(requestBody)
        });

        const result = await response.json();
        showResult(result.valid);
    } catch (error) {
        showError('验证失败: ' + error.message);
    }
}

function showResult(success) {
    resultSection.style.display = 'block';
    resultCard.className = 'result-card ' + (success ? 'success' : 'error');
    resultIcon.textContent = success ? '✅' : '❌';
    resultMessage.textContent = success ? '验证成功！' : '验证失败！';

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
    selectedIndexes = [];
    window.sliderData = null;
    captchaDisplay.innerHTML = '<p class="placeholder">点击下方按钮生成验证码</p>';
    inputSection.style.display = 'none';
    refreshBtn.style.display = 'none';
    hideResult();

    // 清空输入
    document.getElementById('captchaCode').value = '';
    updateSelectedCount();
}
