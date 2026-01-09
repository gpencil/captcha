package captcha

import "errors"

var (
	// ErrCaptchaNotFound 验证码不存在或已过期
	ErrCaptchaNotFound = errors.New("captcha not found or expired")

	// ErrCaptchaInvalid 验证码无效
	ErrCaptchaInvalid = errors.New("captcha invalid")

	// ErrCaptchaExpired 验证码已过期
	ErrCaptchaExpired = errors.New("captcha expired")

	// ErrCaptchaTypeNotSupported 不支持的验证码类型
	ErrCaptchaTypeNotSupported = errors.New("captcha type not supported")

	// ErrCaptchaAnswerFormatWrong 答案格式错误
	ErrCaptchaAnswerFormatWrong = errors.New("captcha answer format wrong")
)
