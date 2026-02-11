// output/markdown.go
package output

import (
	"finance-news/summarizer"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StreamWriter struct {
	path string
	mu   sync.Mutex
	f    *os.File
}

func GetAvailableFilename(fullPath string) (string, error) {
	dir := filepath.Dir(fullPath)                       // 提取目录（如 "./
	fileName := filepath.Base(fullPath)                 // 提取基础文件名（如 "test.txt"）
	ext := filepath.Ext(fileName)                       // 提取扩展名（如 ".txt"）
	nameWithoutExt := fileName[:len(fileName)-len(ext)] // 提取无扩展名的文件名（如 "test"）

	// 2. 先检测原始文件名是否可用
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			// 原始文件不存在，直接返回
			return fullPath, nil
		}
		// 非「不存在」的错误（如权限问题），返回错误
		return "", fmt.Errorf("检测文件状态失败：%w", err)
	}
	// 3. 原始文件存在，循环生成带数字的文件名，直到找到可用的
	counter := 1
	for {
		// 拼接新文件名：名称 + 数字 + 扩展名（如 "test1.txt"）
		newFileName := fmt.Sprintf("%s%d%s", nameWithoutExt, counter, ext)
		// 拼接完整路径
		newFullPath := filepath.Join(dir, newFileName)
		// 检测新路径是否存在
		if _, err := os.Stat(newFullPath); err != nil {
			if os.IsNotExist(err) {
				// 找到可用文件名，返回
				return newFullPath, nil
			}
			// 其他错误，终止循环并返回
			return "", fmt.Errorf("检测文件状态失败：%w", err)
		}
		counter++
		if counter > 999 {
			return "", fmt.Errorf("尝试次数超过上限（999），未找到可用文件名")
		}
	}
}
func NewStreamWriter(outputDir string) (*StreamWriter, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败：%w", err)
	}
	now := time.Now()
	filename := now.Format("2006-01-02") + "_finance_digest.md"
	path := filepath.Join(outputDir, filename)
	availablePath, err := GetAvailableFilename(path)
	if err != nil {
		return nil, fmt.Errorf("获取可用文件名失败：%w", err)
	}
	f, err := os.Create(availablePath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败：%w", err)
	}
	title := fmt.Sprintf("# 金融简报 - %s\n\n", now.Format("2006年1月2日"))
	f.WriteString(title)
	return &StreamWriter{path: availablePath, f: f}, nil
}

func (w *StreamWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.f == nil {
		return nil
	}
	if err := w.f.Close(); err != nil {
		log.Printf("⚠ 关闭文件失败：%v", err)
		return err
	}
	w.f = nil
	return nil
}

func (w *StreamWriter) Write(a summarizer.FinalArticle) {
	block := fmt.Sprintf(`## %s

	
原文**: [%s]


**发布时间**: %s

**摘要**:
%s

---


`, a.TitleZh, a.URL, a.PublishedAt, a.Summary)
	w.f.WriteString(block)
}

func SaveAsMarkdown(now time.Time, articles []summarizer.FinalArticle, outputDir string) {
	os.MkdirAll(outputDir, 0755)

	filename := now.Format("2006-01-02") + "_finance_digest.md"
	path := filepath.Join(outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	title := fmt.Sprintf("# 金融简报 - %s\n\n", now.Format("2006年1月2日"))
	f.WriteString(title)

	for i, a := range articles {
		f.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, a.TitleZh))
		f.WriteString(fmt.Sprintf("**原文**: [%s](%s)\n\n", a.URL, a.URL))
		f.WriteString(fmt.Sprintf("**发布时间**: %s\n\n", a.PublishedAt))
		f.WriteString("**摘要**:\n\n")
		f.WriteString(a.Summary)
		f.WriteString("\n\n---\n\n")
	}

	fmt.Printf("✅ 报告已保存至: %s\n", path)
}
