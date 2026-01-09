package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

// Service 验证码服务
type Service struct {
	store             Store
	characterConfig   CharacterConfig
	imageSelectConfig ImageSelectConfig
	slideConfig       SlideConfig
}

// NewService 创建验证码服务
func NewService(store Store, characterConfig CharacterConfig, imageSelectConfig ImageSelectConfig, slideConfig SlideConfig) *Service {
	return &Service{
		store:             store,
		characterConfig:   characterConfig,
		imageSelectConfig: imageSelectConfig,
		slideConfig:       slideConfig,
	}
}

// Generate 生成验证码
func (s *Service) Generate(ctx context.Context, captchaType CaptchaType) (*CaptchaResponse, error) {
	captchaID := uuid.New().String()

	var data interface{}
	var expireTime time.Duration
	var captchaData interface{}

	switch captchaType {
	case CaptchaTypeCharacter:
		charCaptcha := NewCharacterCaptcha(s.characterConfig)
		code, image, err := charCaptcha.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate character captcha: %w", err)
		}

		data = CharacterData{
			Code: code,
		}
		expireTime = s.characterConfig.ExpireTime
		captchaData = CharacterCaptchaData{
			Image: image,
		}

	case CaptchaTypeImageSelect:
		imageCaptcha := NewImageSelectCaptcha(s.imageSelectConfig)
		question, images, targetIndexes, err := imageCaptcha.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate image select captcha: %w", err)
		}

		data = ImageSelectData{
			TargetIndexes: targetIndexes,
			Question:      question,
		}
		expireTime = s.imageSelectConfig.ExpireTime
		captchaData = ImageSelectCaptchaData{
			Question:    question,
			TargetType:  "bus", // 固定为公交车，实际可根据配置变化
			Images:      images,
			SelectCount: s.imageSelectConfig.SelectCount,
		}

	case SlideTypeSelect:
		slideCaptcha := NewSlideCaptcha(s.slideConfig)
		background, template, _, targetX, err := slideCaptcha.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate slide captcha: %w", err)
		}

		data = SlideData{
			TargetX: targetX,
		}
		expireTime = s.slideConfig.ExpireTime
		captchaData = SlideCaptchaData{
			BackgroundImage: background,
			TemplateImage:   template,
			TemplateY:       0, // 简化实现
			Width:           s.slideConfig.Width,
			Height:          s.slideConfig.Height,
		}

	default:
		return nil, ErrCaptchaTypeNotSupported
	}
	marshal, _ := json.Marshal(data)
	fmt.Println("存储的验证码", string(marshal))

	// 存储验证码数据
	err := s.store.Set(ctx, captchaID, data, expireTime)
	if err != nil {
		logx.Errorf("failed to store captcha: %v", err)
		return nil, fmt.Errorf("failed to store captcha: %w", err)
	}

	return &CaptchaResponse{
		CaptchaID:   captchaID,
		CaptchaType: captchaType,
		Data:        captchaData,
		ExpireTime:  time.Now().Add(expireTime).Unix(),
	}, nil
}

// Verify 验证验证码
func (s *Service) Verify(ctx context.Context, req *VerifyRequest) (bool, error) {
	// 获取存储的验证码数据
	value, err := s.store.Get(ctx, req.CaptchaID)
	if err != nil {
		fmt.Println("拿到错误了", err)
		if err == ErrCaptchaNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get captcha: %w", err)
	}

	// 根据类型验证
	switch req.CaptchaType {
	case CaptchaTypeCharacter:
		return s.verifyCharacter(value, req.Answer)
	case CaptchaTypeImageSelect:
		return s.verifyImageSelect(value, req.Answer)
	case SlideTypeSelect:
		return s.verifySlide(value, req.Answer)
	default:
		return false, ErrCaptchaTypeNotSupported
	}
}

// VerifyAndDelete 验证并删除验证码
func (s *Service) VerifyAndDelete(ctx context.Context, req *VerifyRequest) (bool, error) {
	valid, err := s.Verify(ctx, req)
	if err != nil {
		return false, err
	}

	if valid {
		// 验证成功，删除验证码
		_ = s.store.Del(ctx, req.CaptchaID)
	}

	return valid, nil
}

// verifyCharacter 验证字符验证码
func (s *Service) verifyCharacter(value string, answer interface{}) (bool, error) {
	// 解析存储的数据
	var data CharacterData
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal captcha data: %w", err)
	}

	// 解析答案
	answerBytes, err := json.Marshal(answer)
	if err != nil {
		return false, ErrCaptchaAnswerFormatWrong
	}

	var answerData CharacterAnswer
	err = json.Unmarshal(answerBytes, &answerData)
	if err != nil {
		return false, ErrCaptchaAnswerFormatWrong
	}

	// 验证
	charCaptcha := NewCharacterCaptcha(s.characterConfig)
	return charCaptcha.Verify(data.Code, answerData.Code), nil
}

// verifyImageSelect 验证图片选择验证码
func (s *Service) verifyImageSelect(value string, answer interface{}) (bool, error) {
	// 解析存储的数据
	var data ImageSelectData
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal captcha data: %w", err)
	}

	// 解析答案
	answerBytes, err := json.Marshal(answer)
	if err != nil {
		return false, ErrCaptchaAnswerFormatWrong
	}

	var answerData ImageSelectAnswer
	err = json.Unmarshal(answerBytes, &answerData)
	if err != nil {
		return false, ErrCaptchaAnswerFormatWrong
	}

	// 验证
	imageCaptcha := NewImageSelectCaptcha(s.imageSelectConfig)
	return imageCaptcha.Verify(data.TargetIndexes, answerData.SelectedIndexes), nil
}

// verifySlide 验证滑动验证码
func (s *Service) verifySlide(value string, answer interface{}) (bool, error) {
	// 解析存储的数据
	var data SlideData
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal captcha data: %w", err)
	}

	// 解析答案
	answerBytes, err := json.Marshal(answer)
	if err != nil {
		return false, ErrCaptchaAnswerFormatWrong
	}

	var answerData SlideAnswer
	err = json.Unmarshal(answerBytes, &answerData)
	if err != nil {
		return false, ErrCaptchaAnswerFormatWrong
	}

	// 验证
	slideCaptcha := NewSlideCaptcha(s.slideConfig)
	return slideCaptcha.Verify(data.TargetX, answerData), nil
}

// CharacterData 字符验证码存储数据
type CharacterData struct {
	Code string `json:"code"`
}

// ImageSelectData 图片选择验证码存储数据
type ImageSelectData struct {
	TargetIndexes []int  `json:"targetIndexes"` // 正确答案的索引
	Question      string `json:"question"`      // 问题
}

// SlideData 滑动验证码存储数据
type SlideData struct {
	TargetX int `json:"targetX"` // 正确的X坐标
}
