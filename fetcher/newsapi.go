// fetcher/newsapi.go
package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
	Content     string `json:"content"`
}

type NewsResponse struct {
	Status   string    `json:"status"`
	Total    int       `json:"totalResults"`
	Articles []Article `json:"articles"`
}

// 将关键词转为 NewsAPI 支持的 OR 查询
func buildQuery(keywords, exclude []string) string {
	var parts []string

	for _, k := range keywords {
		parts = append(parts, `"`+strings.TrimSpace(k)+`"`)
	}

	q := "(" + strings.Join(parts, " OR ") + ")"

	if len(exclude) > 0 {
		var ex []string
		for _, e := range exclude {
			ex = append(ex, e)
		}
		q += " NOT (" + strings.Join(ex, " OR ") + ")"
	}

	return q
}

func FetchRelevantNews(client *http.Client, apiKey string, keywords []string) ([]Article, error) {
	exclude := []string{
		"sport", "sports", "football", "soccer",
		"nba", "olympic", "medal", "player",
		"team", "match", "tournament",
	}

	query := buildQuery(keywords, exclude)
	baseURL := "https://newsapi.org/v2/everything"
	params := url.Values{}
	params.Add("q", query)
	params.Add("language", "en")
	params.Add("sortBy", "publishedAt")
	params.Add("pageSize", "30")
	params.Add("apiKey", apiKey)
	fullURL := baseURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 NewsAPI 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("NewsAPI 返回非200状态码: %d", resp.StatusCode)
	}

	var newsResp NewsResponse
	err = json.NewDecoder(resp.Body).Decode(&newsResp)
	if err != nil {
		return nil, fmt.Errorf("解析 NewsAPI 响应失败: %w", err)
	}

	if newsResp.Status != "ok" {
		return nil, fmt.Errorf("NewsAPI 错误: total=%d", newsResp.Total)
	}

	return newsResp.Articles, nil
}
