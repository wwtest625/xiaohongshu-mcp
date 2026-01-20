package browser

import (
	"encoding/json"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

// Browser 封装了原有的 headless_browser 功能，并支持禁用 leakless
type Browser struct {
	browser  *rod.Browser
	launcher *launcher.Launcher
}

type browserConfig struct {
	headless      bool
	binPath       string
	userAgent     string
	cookies       string
}

type Option func(*browserConfig)

func WithBinPath(binPath string) Option {
	return func(c *browserConfig) {
		c.binPath = binPath
	}
}

func WithHeadless(headless bool) Option {
	return func(c *browserConfig) {
		c.headless = headless
	}
}

func NewBrowser(headless bool, options ...Option) *Browser {
	cfg := &browserConfig{
		headless:  headless,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
	for _, opt := range options {
		opt(cfg)
	}

	// 加载 cookies
	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)
	if data, err := cookieLoader.LoadCookies(); err == nil {
		cfg.cookies = string(data)
		logrus.Debugf("已从文件加载 cookies")
	}

	// 配置 Launcher
	l := launcher.New().
		Headless(cfg.headless).
		Leakless(false). // 彻底禁用引起报毒的 leakless
		Set("--no-sandbox")

	if cfg.binPath != "" {
		l = l.Bin(cfg.binPath)
	}

	l.Set("user-agent", cfg.userAgent)

	url := l.MustLaunch()

	b := rod.New().
		ControlURL(url).
		MustConnect()

	// 设置 cookies
	if cfg.cookies != "" {
		var cks []*proto.NetworkCookie
		if err := json.Unmarshal([]byte(cfg.cookies), &cks); err == nil {
			b.MustSetCookies(cks...)
		}
	}

	return &Browser{
		browser:  b,
		launcher: l,
	}
}

func (b *Browser) Close() {
	if b.browser != nil {
		_ = b.browser.Close()
	}
	if b.launcher != nil {
		b.launcher.Cleanup()
	}
}

func (b *Browser) NewPage() *rod.Page {
	return stealth.MustPage(b.browser)
}
