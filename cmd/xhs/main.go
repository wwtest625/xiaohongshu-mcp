package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

func main() {
	var (
		action    string
		feedID    string
		token     string
		keyword   string
		output    string
		isHeadless bool
		title      string
		content    string
		tags       string
		images     string
		video      string
	)

	flag.StringVar(&action, "action", "fetch", "操作类型: fetch, search, publish_image, publish_video, check_login, recommend")
	flag.StringVar(&feedID, "id", "", "笔记 ID")
	flag.StringVar(&token, "token", "", "xsec_token")
	flag.StringVar(&keyword, "keyword", "", "搜索关键词")
	flag.StringVar(&output, "output", "result.json", "结果输出文件")
	flag.BoolVar(&isHeadless, "headless", true, "是否无头模式")
	flag.StringVar(&title, "title", "", "发布标题")
	flag.StringVar(&content, "content", "", "发布正文")
	flag.StringVar(&tags, "tags", "", "标签(逗号分隔)")
	flag.StringVar(&images, "images", "", "图片路径(逗号分隔)")
	flag.StringVar(&video, "video", "", "视频路径")
	flag.Parse()

	// 强制设置 ROD_LEAKLESS 为 false 修复 Windows 兼容性
	os.Setenv("ROD_LEAKLESS", "false")

	logrus.SetLevel(logrus.InfoLevel)

	switch action {
	case "fetch":
		if feedID == "" || token == "" {
			fmt.Println("错误: fetch 操作需要提供 -id 和 -token")
			os.Exit(1)
		}
		fetchComments(feedID, token, output, isHeadless)
	case "search":
		if keyword == "" {
			fmt.Println("错误: search 操作需要提供 -keyword")
			os.Exit(1)
		}
		searchFeeds(keyword, output, isHeadless)
	case "publish_image":
		if title == "" || content == "" || images == "" {
			fmt.Println("错误: publish_image 需要 -title, -content 和 -images")
			os.Exit(1)
		}
		publishImage(title, content, strings.Split(tags, ","), strings.Split(images, ","), isHeadless)
	case "publish_video":
		if title == "" || content == "" || video == "" {
			fmt.Println("错误: publish_video 需要 -title, -content 和 -video")
			os.Exit(1)
		}
		publishVideo(title, content, strings.Split(tags, ","), video, isHeadless)
	case "check_login":
		checkLogin(isHeadless)
	case "recommend":
		getRecommend(output, isHeadless)
	default:
		fmt.Printf("未知操作: %s\n", action)
		os.Exit(1)
	}
}

func fetchComments(feedID, token, output string, headless bool) {
	b := browser.NewBrowser(headless)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := xiaohongshu.NewFeedDetailAction(page)
	config := xiaohongshu.DefaultCommentLoadConfig()
	
	result, err := action.GetFeedDetailWithConfig(context.Background(), feedID, token, true, config)
	
	if err != nil {
		logrus.Errorf("抓取失败: %v", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	if err := os.WriteFile(output, data, 0644); err != nil {
		logrus.Errorf("保存失败: %v", err)
		os.Exit(1)
	}
	fmt.Printf("抓取成功，数据已保存至: %s\n", output)
}

func searchFeeds(keyword, output string, headless bool) {
	b := browser.NewBrowser(headless)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := xiaohongshu.NewSearchAction(page)
	feeds, err := action.Search(context.Background(), keyword)
	if err != nil {
		logrus.Errorf("搜索失败: %v", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(feeds, "", "  ")
	if err := os.WriteFile(output, data, 0644); err != nil {
		logrus.Errorf("保存失败: %v", err)
		os.Exit(1)
	}
	fmt.Printf("搜索成功，数据已保存至: %s\n", output)
}

func publishImage(title, content string, tags, images []string, headless bool) {
	b := browser.NewBrowser(headless)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action, err := xiaohongshu.NewPublishImageAction(page)
	if err != nil {
		logrus.Errorf("进入发布页失败: %v", err)
		os.Exit(1)
	}

	payload := xiaohongshu.PublishImageContent{
		Title:      title,
		Content:    content,
		Tags:       tags,
		ImagePaths: images,
	}

	if err := action.Publish(context.Background(), payload); err != nil {
		logrus.Errorf("发布失败: %v", err)
		os.Exit(1)
	}
	fmt.Println("🎉 图文笔记发布成功！")
}

func publishVideo(title, content string, tags []string, video string, headless bool) {
	b := browser.NewBrowser(headless)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action, err := xiaohongshu.NewPublishVideoAction(page)
	if err != nil {
		logrus.Errorf("进入发布页失败: %v", err)
		os.Exit(1)
	}

	payload := xiaohongshu.PublishVideoContent{
		Title:     title,
		Content:   content,
		Tags:      tags,
		VideoPath: video,
	}

	if err := action.PublishVideo(context.Background(), payload); err != nil {
		logrus.Errorf("发布失败: %v", err)
		os.Exit(1)
	}
	fmt.Println("🎉 视频笔记发布成功！")
}

func checkLogin(headless bool) {
	b := browser.NewBrowser(headless)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := xiaohongshu.NewLogin(page)
	isLogin, err := action.CheckLoginStatus(context.Background())
	if err != nil {
		fmt.Printf("登录状态检查失败: %v\n", err)
		os.Exit(1)
	}

	if isLogin {
		fmt.Println("STATUS: LOGGED_IN")
	} else {
		fmt.Println("STATUS: NOT_LOGGED_IN")
		os.Exit(2) // 使用特定退出码表示未登录
	}
}

func getRecommend(output string, headless bool) {
	b := browser.NewBrowser(headless)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := xiaohongshu.NewFeedsListAction(page)
	feeds, err := action.GetFeedsList(context.Background())
	if err != nil {
		logrus.Errorf("获取推荐列表失败: %v", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(feeds, "", "  ")
	if err := os.WriteFile(output, data, 0644); err != nil {
		logrus.Errorf("保存结果失败: %v", err)
		os.Exit(1)
	}
	fmt.Printf("推荐列表获取成功，数据已保存至: %s\n", output)
}
