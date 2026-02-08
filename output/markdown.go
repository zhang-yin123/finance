// output/markdown.go
package output

import (
	"finance-news/summarizer"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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
