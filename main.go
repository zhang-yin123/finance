// main.go
package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"finance-news/config"
	"finance-news/fetcher"
	"finance-news/output"
	"finance-news/summarizer"
)

func NewHTTPClient(proxyAddr string) *http.Client {
	client := &http.Client{
		Timeout: time.Minute * 5,
	}
	if proxyAddr != "" {
		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			log.Printf("代理地址无效: %v", err)
			return client
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	return client
}

func concurrencyReport() {
	cfg := config.LoadConfig("config.yaml")
	client := NewHTTPClient(cfg.ClashProxy)
	// 1. 抓新闻
	articles, err := fetcher.FetchRelevantNews(client, cfg.NewsAPIKey, cfg.Keywords)
	if err != nil {
		log.Fatal("抓取新闻失败:", err)
	}
	log.Printf("成功获取 %d 篇相关新闻", len(articles))

	writer, err := output.NewStreamWriter(cfg.OutputDir)
	if err != nil {
		log.Fatal("初始化写入器失败:", err)
	}
	defer writer.Close()
	s := summarizer.NewDeepSeek(cfg.DeepSeekAPIKey)
	ch := make(chan summarizer.FinalArticle, len(articles))
	s.SummarizeConcurrent(articles, ch)
	for val := range ch {
		writer.Write(val)
	}
	log.Println("✅ 日报生成完成")
}

func serialReport() {
	cfg := config.LoadConfig("config.yaml")

	client := NewHTTPClient(cfg.ClashProxy)

	// 1. 抓新闻
	articles, err := fetcher.FetchRelevantNews(client, cfg.NewsAPIKey, cfg.Keywords)
	if err != nil {
		log.Fatal("抓取新闻失败:", err)
	}
	log.Printf("成功获取 %d 篇相关新闻", len(articles))
	// 2. ai总结
	s := summarizer.NewDeepSeek(cfg.DeepSeekAPIKey)
	final := s.Summarize(articles)
	// 4. 输出Markdown
	output.SaveAsMarkdown(time.Now(), final, cfg.OutputDir)

	log.Println("✅ 日报生成完成")
}

func main() {
	concurrencyReport()
}
