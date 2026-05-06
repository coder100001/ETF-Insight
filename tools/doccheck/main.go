package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/coder100001/etf-insight/tools/doccheck/checker"
	"github.com/coder100001/etf-insight/tools/doccheck/models"
	"github.com/coder100001/etf-insight/tools/doccheck/reporter"
)

func main() {
	projectRoot := flag.String("project-root", ".", "项目根目录路径")
	output := flag.String("output", "", "报告输出文件路径")
	format := flag.String("format", "markdown", "报告格式: markdown | json")
	strict := flag.Bool("strict", false, "严格模式：存在高严重性问题时返回非零退出码")
	quick := flag.Bool("quick", false, "快速模式：仅检查变更文件关联的元素")
	changedFiles := flag.String("changed-files", "", "变更文件列表（逗号分隔）")
	flag.Parse()

	absRoot, err := absPath(*projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 无效的项目路径: %v\n", err)
		os.Exit(1)
	}

	if *quick {
		fmt.Println("🔍 开始文档-代码一致性检查（快速模式）...")
	} else {
		fmt.Println("🔍 开始文档-代码一致性检查...")
	}
	fmt.Printf("   项目路径: %s\n\n", absRoot)

	c := checker.NewConsistencyChecker(absRoot)

	var result *models.CheckResult
	if *quick {
		// 快速模式
		files := []string{}
		if *changedFiles != "" {
			files = strings.Split(*changedFiles, ",")
		} else if envFiles := os.Getenv("DOCCHECK_CHANGED_FILES"); envFiles != "" {
			files = strings.Split(envFiles, ",")
		}
		qc := checker.NewQuickChecker(c, files)
		result, err = qc.Run()
	} else {
		// 完整模式
		result, err = c.Run()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 检查失败: %v\n", err)
		os.Exit(1)
	}

	rep := reporter.NewReporter(absRoot)

	if *output != "" {
		if err := rep.WriteReport(result, *output, *format); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 写入报告失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("📄 报告已保存到: %s\n\n", *output)
	}

	switch *format {
	case "json":
		fmt.Println(rep.GenerateJSON(result))
	default:
		fmt.Println(rep.GenerateMarkdown(result))
	}

	if *strict {
		highCount := result.IssuesBySeverity["high"]
		if highCount > 0 {
			fmt.Fprintf(os.Stderr, "\n❌ 发现 %d 个高严重性问题，请先修复\n", highCount)
			os.Exit(1)
		}
	}

	if result.ConsistencyScore < 60 {
		fmt.Fprintf(os.Stderr, "\n⚠️ 一致性得分 %.1f 低于60分，建议尽快改善文档\n", result.ConsistencyScore)
	}
}

func absPath(path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		return path, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd + "/" + path, nil
}
