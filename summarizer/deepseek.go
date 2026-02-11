package summarizer

import (
	"bytes"
	"encoding/json"
	"finance-news/fetcher"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var truncationRegex = regexp.MustCompile(`\s*\[\+\d+\s+chars\]$`)

type FinalArticle struct {
	TitleZh       string
	DescriptionZh string
	URL           string
	PublishedAt   string
	Summary       string
}

type DeepSeek struct {
	APIKey string
}

func NewDeepSeek(key string) *DeepSeek {
	return &DeepSeek{APIKey: key}
}

func cleanContent(content string) string {
	content = strings.TrimSpace(content)
	return truncationRegex.ReplaceAllString(content, "")
}

/** 并发总结*/
func (s *DeepSeek) SummarizeConcurrent(articles []fetcher.Article, ch chan FinalArticle) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 30) // 限制最大并发30个，避免API限流
	for i, a := range articles {
		wg.Add(1)
		go func(idx int, art fetcher.Article) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量
			sem <- struct{}{}        // 获取信号量
			desc := strings.TrimSpace(art.Description)
			content := cleanContent(art.Content)
			prompt := fmt.Sprintf(`你是专业宏观金融分析师，请基于以下新闻完成任务：
1. 提炼并翻译成中文新闻要点
2. 核心结论
3. 对美元影响（上涨/下跌/中性）
4. 对黄金影响
5. 风险等级（低/中/高）
新闻摘要:
%s
新闻正文:
%s`, desc, content)
			log.Printf("🚀 调用大模型 %d\n", idx+1)
			summary := s.callDeepSeek(prompt)
			ch <- FinalArticle{
				TitleZh:       fmt.Sprintf("%d（模型生成）", idx+1),
				DescriptionZh: "（模型生成）",
				URL:           art.URL,
				PublishedAt:   art.PublishedAt,
				Summary:       summary,
			}
		}(i, a)
	}

	wg.Wait()
	close(ch)
}

func (s *DeepSeek) callDeepSeek(prompt string) string {
	// ✅ 使用 /v1 路径（官方兼容 OpenAI）
	apiURL := "https://api.deepseek.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.2,
		"max_tokens":  600,
		"stream":      false,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "【摘要失败】网络错误: " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Sprintf("【摘要失败】HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(resp.Body).Decode(&apiResp)

	if len(apiResp.Choices) == 0 {
		return "【摘要失败】无返回内容"
	}
	return strings.TrimSpace(apiResp.Choices[0].Message.Content)
}

/** 批量串行总结 */
func (s *DeepSeek) Summarize(articles []fetcher.Article) []FinalArticle {
	var result []FinalArticle

	for i, a := range articles {
		desc := strings.TrimSpace(a.Description)
		content := cleanContent(a.Content)
		prompt := fmt.Sprintf(`你是专业宏观金融分析师，请基于以下新闻完成任务：
1. 提炼并翻译成中文新闻要点
2. 核心结论
3. 对美元影响（上涨/下跌/中性）
4. 对黄金影响
5. 风险等级（低/中/高）
新闻摘要:
%s
新闻正文:
%s`, desc, content)
		log.Printf("调用大模型 %d 次 \n", i+1)
		summary := s.callDeepSeek(prompt)

		result = append(result, FinalArticle{
			TitleZh:       "（模型生成）",
			DescriptionZh: "（模型生成）",
			URL:           a.URL,
			PublishedAt:   a.PublishedAt,
			Summary:       summary,
		})
	}

	return result
}
