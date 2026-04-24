package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coder100001/etf-insight/tools/doccheck/models"
)

type Reporter struct {
	projectRoot string
}

func NewReporter(projectRoot string) *Reporter {
	return &Reporter{projectRoot: projectRoot}
}

func (r *Reporter) GenerateMarkdown(result *models.CheckResult) string {
	var b strings.Builder

	b.WriteString("# 文档-代码一致性检查报告\n\n")

	b.WriteString(fmt.Sprintf("- **生成时间**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("- **项目路径**: %s\n\n", r.projectRoot))

	b.WriteString("## 📊 检查摘要\n\n")

	b.WriteString("| 指标 | 数值 |\n")
	b.WriteString("|------|------|\n")
	b.WriteString(fmt.Sprintf("| 一致性得分 | %.1f/100 |\n", result.ConsistencyScore))
	b.WriteString(fmt.Sprintf("| 问题总数 | %d |\n", result.TotalIssues))
	b.WriteString(fmt.Sprintf("| 代码元素数 | %d |\n", result.CodeElementsCount))
	b.WriteString(fmt.Sprintf("| 文档章节数 | %d |\n\n", result.DocSectionsCount))

	b.WriteString("### 按严重程度\n\n")
	b.WriteString("| 严重程度 | 数量 |\n")
	b.WriteString("|----------|------|\n")
	for _, sev := range []string{"high", "medium", "low"} {
		if count, ok := result.IssuesBySeverity[sev]; ok {
			b.WriteString(fmt.Sprintf("| %s | %d |\n", strings.ToUpper(sev), count))
		}
	}
	b.WriteString("\n")

	b.WriteString("### 按问题类型\n\n")
	b.WriteString("| 问题类型 | 数量 |\n")
	b.WriteString("|----------|------|\n")
	for typ, count := range result.IssuesByType {
		b.WriteString(fmt.Sprintf("| %s | %d |\n", typ, count))
	}
	b.WriteString("\n")

	if len(result.Issues) > 0 {
		b.WriteString("## 🔍 问题详情\n\n")

		highIssues := filterBySeverity(result.Issues, "high")
		mediumIssues := filterBySeverity(result.Issues, "medium")

		if len(highIssues) > 0 {
			b.WriteString("### 🔴 高严重性问题\n\n")
			for i, issue := range highIssues {
				b.WriteString(fmt.Sprintf("%d. **[%s]** %s\n", i+1, issue.Type, issue.Message))
				if issue.Document != "" {
					b.WriteString(fmt.Sprintf("   - 文档: %s", issue.Document))
					if issue.DocumentLine > 0 {
						b.WriteString(fmt.Sprintf(":%d", issue.DocumentLine))
					}
					b.WriteString("\n")
				}
				if issue.ElementFile != "" {
					b.WriteString(fmt.Sprintf("   - 代码: %s\n", issue.ElementFile))
				}
				b.WriteString(fmt.Sprintf("   - 💡 建议: %s\n\n", r.suggestFix(issue)))
			}
		}

		if len(mediumIssues) > 0 {
			b.WriteString("### 🟡 中等严重性问题\n\n")
			maxShow := 30
			for i, issue := range mediumIssues {
				if i >= maxShow {
					b.WriteString(fmt.Sprintf("\n... 还有 %d 个中等严重性问题未显示\n\n", len(mediumIssues)-maxShow))
					break
				}
				b.WriteString(fmt.Sprintf("%d. **[%s]** %s\n", i+1, issue.Type, issue.Message))
				if issue.ElementFile != "" {
					b.WriteString(fmt.Sprintf("   - 代码: %s\n", issue.ElementFile))
				}
				b.WriteString(fmt.Sprintf("   - 💡 建议: %s\n\n", r.suggestFix(issue)))
			}
		}
	} else {
		b.WriteString("## ✅ 未发现不一致问题\n\n")
	}

	b.WriteString("## 💡 改进建议\n\n")
	b.WriteString(r.generateRecommendations(result))
	b.WriteString("\n")

	return b.String()
}

func (r *Reporter) GenerateJSON(result *models.CheckResult) string {
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

func (r *Reporter) WriteReport(result *models.CheckResult, outputPath string, format string) error {
	var content string
	switch format {
	case "json":
		content = r.GenerateJSON(result)
	default:
		content = r.GenerateMarkdown(result)
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func (r *Reporter) suggestFix(issue models.Issue) string {
	switch issue.Type {
	case "undocumented_code_element":
		return fmt.Sprintf("在相关文档中添加对 '%s' 的说明", issue.ElementName)
	case "version_mismatch":
		return fmt.Sprintf("统一版本号为 %s 或 %s", issue.ReadmeVersion, issue.CodeVersion)
	case "unimplemented_feature":
		return fmt.Sprintf("实现 '%s' 功能，或从文档中删除相关描述", issue.Feature)
	case "undocumented_feature":
		return fmt.Sprintf("在文档中添加对 '%s' 功能的说明", issue.Feature)
	case "missing_required_document":
		return fmt.Sprintf("创建必需的文档文件 '%s'", issue.Document)
	default:
		return "请检查并更新相关文档"
	}
}

func (r *Reporter) generateRecommendations(result *models.CheckResult) string {
	var recs []string

	if count, ok := result.IssuesBySeverity["high"]; ok && count > 0 {
		recs = append(recs, fmt.Sprintf("- **立即处理 %d 个高严重性问题**：这些问题可能导致用户无法使用文档中描述的功能", count))
	}

	if count, ok := result.IssuesBySeverity["medium"]; ok && count > 0 {
		recs = append(recs, fmt.Sprintf("- **尽快处理 %d 个中等严重性问题**：这些问题影响用户体验", count))
	}

	if result.ConsistencyScore < 80 {
		recs = append(recs, "- **提升文档覆盖率**：当前一致性得分低于80分，需要补充更多文档")
	}

	recs = append(recs, "- **建立文档更新流程**：代码变更时同步更新相关文档")
	recs = append(recs, "- **配置Git预提交钩子**：`make doccheck-hook` 启用自动检查")

	return strings.Join(recs, "\n")
}

func filterBySeverity(issues []models.Issue, severity string) []models.Issue {
	var filtered []models.Issue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}