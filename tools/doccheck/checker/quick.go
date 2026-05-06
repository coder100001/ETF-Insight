package checker

import (
	"fmt"
	"strings"

	"github.com/coder100001/etf-insight/tools/doccheck/models"
)

// QuickChecker 快速模式检查器 - 只检查变更文件关联的元素
type QuickChecker struct {
	base         *ConsistencyChecker
	changedFiles []string
}

// NewQuickChecker 创建快速检查器
func NewQuickChecker(base *ConsistencyChecker, changedFiles []string) *QuickChecker {
	return &QuickChecker{
		base:         base,
		changedFiles: changedFiles,
	}
}

// Run 执行快速检查
func (qc *QuickChecker) Run() (*models.CheckResult, error) {
	c := qc.base

	// 先执行基础扫描
	if err := c.codeScanner.Scan(); err != nil {
		return nil, fmt.Errorf("代码扫描失败: %w", err)
	}
	if err := c.docParser.Parse(); err != nil {
		return nil, fmt.Errorf("文档解析失败: %w", err)
	}
	c.loadMapping()
	c.issues = c.issues[:0]

	// 1. 提取受影响的代码元素
	affectedElements := qc.extractAffectedElements()

	// 2. 检查这些元素的文档覆盖（代码→文档）
	for _, elem := range affectedElements {
		if !c.isNameDocumented(elem.Name) {
			c.issues = append(c.issues, models.Issue{
				Type:        "undocumented_code_element",
				Severity:    "medium",
				Message:     fmt.Sprintf("代码元素 '%s' 在文档中未提及", elem.Name),
				ElementName: elem.Name,
				ElementType: elem.Type,
			})
		}
	}

	// 3. 检查文档中提到的相关功能是否已实现（文档→代码）
	qc.checkRelatedFeatures(affectedElements)

	// 4. 版本一致性（每次必查）
	c.checkVersionConsistency()

	// 5. 必需文档存在性
	c.checkRequiredDocuments()

	return c.buildResult(), nil
}

// extractAffectedElements 从变更文件提取受影响的代码元素
func (qc *QuickChecker) extractAffectedElements() []models.CodeElement {
	c := qc.base
	var affected []models.CodeElement

	for _, elem := range c.codeScanner.Elements() {
		for _, changedFile := range qc.changedFiles {
			changedFile = strings.TrimSpace(changedFile)
			if changedFile == "" {
				continue
			}
			// 检查元素文件路径是否匹配变更文件
			if strings.HasSuffix(elem.FilePath, changedFile) ||
				strings.HasSuffix(changedFile, elem.FilePath) ||
				strings.Contains(elem.FilePath, changedFile) {
				affected = append(affected, elem)
				break
			}
		}
	}

	return affected
}

// checkRelatedFeatures 检查与变更元素相关的功能映射
func (qc *QuickChecker) checkRelatedFeatures(elements []models.CodeElement) {
	c := qc.base
	if c.mapping == nil {
		return
	}

	// 构建变更元素的名称集合
	elementNames := make(map[string]bool)
	for _, elem := range elements {
		baseName := extractBaseName(elem.Name)
		elementNames[baseName] = true
		elementNames[elem.Name] = true
	}

	// 只检查与变更元素相关的功能
	for featureName, mapping := range c.mapping.FeatureMappings {
		if !qc.isFeatureRelated(featureName, mapping, elementNames) {
			continue
		}

		codeExists := c.featureExistsInCode(mapping.CodeIndicators)
		docExists := c.featureExistsInDocs(mapping.DocumentSections)

		if docExists && !codeExists {
			c.issues = append(c.issues, models.Issue{
				Type:     "unimplemented_feature",
				Severity: "high",
				Message:  fmt.Sprintf("文档中提到的功能 '%s' 在代码中未实现", featureName),
				Feature:  featureName,
			})
		}
	}
}

// isFeatureRelated 检查功能是否与变更元素相关
func (qc *QuickChecker) isFeatureRelated(featureName string, mapping models.FeatureMapping, elementNames map[string]bool) bool {
	if elementNames[featureName] {
		return true
	}
	for _, indicator := range mapping.CodeIndicators {
		if elementNames[indicator] {
			return true
		}
	}
	return false
}

// extractBaseName 提取基础名称（去掉常见后缀）
func extractBaseName(name string) string {
	base := name
	suffixes := []string{"Handler", "Service", "Controller", "Model", "Repository", "Manager"}
	for _, suffix := range suffixes {
		base = strings.TrimSuffix(base, suffix)
	}
	return base
}
