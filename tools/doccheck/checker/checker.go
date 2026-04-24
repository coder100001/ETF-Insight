package checker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coder100001/etf-insight/tools/doccheck/models"
	"github.com/coder100001/etf-insight/tools/doccheck/parser"
	"github.com/coder100001/etf-insight/tools/doccheck/scanner"
)

var (
	versionPattern    = regexp.MustCompile(`v(\d+\.\d+\.\d+)`)
	pkgVersionPattern = regexp.MustCompile(`"version"\s*:\s*"([^"]+)"`)
)

type ConsistencyChecker struct {
	projectRoot string
	codeScanner *scanner.CodeScanner
	docParser   *parser.DocumentParser
	issues      []models.Issue
	mapping     *models.MappingConfig
}

func NewConsistencyChecker(projectRoot string) *ConsistencyChecker {
	return &ConsistencyChecker{
		projectRoot: projectRoot,
		codeScanner: scanner.NewCodeScanner(projectRoot),
		docParser:   parser.NewDocumentParser(projectRoot),
		issues:      make([]models.Issue, 0),
	}
}

func (c *ConsistencyChecker) Issues() []models.Issue {
	return c.issues
}

func (c *ConsistencyChecker) Run() (*models.CheckResult, error) {
	if err := c.codeScanner.Scan(); err != nil {
		return nil, fmt.Errorf("代码扫描失败: %w", err)
	}

	if err := c.docParser.Parse(); err != nil {
		return nil, fmt.Errorf("文档解析失败: %w", err)
	}

	c.loadMapping()
	c.issues = c.issues[:0]

	c.checkUndocumentedElements()
	c.checkVersionConsistency()
	c.checkFeatureConsistency()
	c.checkRequiredDocuments()

	result := c.buildResult()
	return result, nil
}

func (c *ConsistencyChecker) loadMapping() {
	mappingFile := filepath.Join(c.projectRoot, "docs", "consistency", "mapping_rules.json")
	data, err := os.ReadFile(mappingFile)
	if err != nil {
		c.mapping = &models.MappingConfig{}
		return
	}

	var config models.MappingConfig
	if err := json.Unmarshal(data, &config); err != nil {
		c.mapping = &models.MappingConfig{}
		return
	}
	c.mapping = &config
}

func (c *ConsistencyChecker) checkUndocumentedElements() {
	handlerAndServiceStructs := make(map[string]bool)
	for _, elem := range c.codeScanner.Elements() {
		if (elem.Type == "handler" || elem.Type == "service") && !strings.Contains(elem.Name, "(") {
			if strings.HasSuffix(elem.Name, "Handler") || strings.HasSuffix(elem.Name, "Service") {
				handlerAndServiceStructs[elem.Name] = false
			}
		}
	}

	for name := range handlerAndServiceStructs {
		if c.isNameDocumented(name) {
			continue
		}
		c.issues = append(c.issues, models.Issue{
			Type:        "undocumented_code_element",
			Severity:    "medium",
			Message:     fmt.Sprintf("重要的代码元素 '%s' 在文档中未提及", name),
			ElementName: name,
			ElementType: "handler/service",
		})
	}

	componentNames := make(map[string]bool)
	for _, elem := range c.codeScanner.Elements() {
		if elem.Type == "component" {
			componentNames[elem.Name] = false
		}
	}

	for name := range componentNames {
		if c.isNameDocumented(name) {
			continue
		}
		c.issues = append(c.issues, models.Issue{
			Type:        "undocumented_code_element",
			Severity:    "medium",
			Message:     fmt.Sprintf("前端组件 '%s' 在文档中未提及", name),
			ElementName: name,
			ElementType: "component",
		})
	}
}

func (c *ConsistencyChecker) isNameDocumented(name string) bool {
	keywords := c.extractKeywords(name)
	if len(keywords) == 0 {
		return false
	}

	for _, section := range c.docParser.Sections() {
		for _, kw := range keywords {
			if strings.Contains(section.Content, kw) || strings.Contains(section.Title, kw) {
				return true
			}
		}
	}
	return false
}

func (c *ConsistencyChecker) extractKeywords(name string) []string {
	base := name
	base = strings.TrimSuffix(base, "Handler")
	base = strings.TrimSuffix(base, "Service")
	base = strings.TrimSuffix(base, "Controller")

	if base == "" {
		return nil
	}

	var keywords []string
	keywords = append(keywords, base)

	if len(base) > 4 {
		parts := camelCaseSplit(base)
		if len(parts) > 1 {
			keywords = append(keywords, parts...)
		}
	}

	return keywords
}

func camelCaseSplit(s string) []string {
	var parts []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func (c *ConsistencyChecker) checkVersionConsistency() {
	readmeVersion := c.extractReadmeVersion()
	codeVersion := c.extractCodeVersion()

	if readmeVersion != "" && codeVersion != "" && readmeVersion != codeVersion {
		c.issues = append(c.issues, models.Issue{
			Type:          "version_mismatch",
			Severity:      "high",
			Message:       fmt.Sprintf("文档版本(%s)与代码版本(%s)不一致", readmeVersion, codeVersion),
			ReadmeVersion: readmeVersion,
			CodeVersion:   codeVersion,
		})
	}
}

func (c *ConsistencyChecker) checkFeatureConsistency() {
	if c.mapping == nil {
		return
	}

	for featureName, mapping := range c.mapping.FeatureMappings {
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

		if codeExists && !docExists && mapping.Required {
			c.issues = append(c.issues, models.Issue{
				Type:     "undocumented_feature",
				Severity: "medium",
				Message:  fmt.Sprintf("代码中实现的功能 '%s' 在文档中未提及", featureName),
				Feature:  featureName,
			})
		}
	}
}

func (c *ConsistencyChecker) checkRequiredDocuments() {
	if c.mapping == nil {
		return
	}

	for _, doc := range c.mapping.ValidationRules.RequiredDocuments {
		path := filepath.Join(c.projectRoot, doc)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			c.issues = append(c.issues, models.Issue{
				Type:     "missing_required_document",
				Severity: "high",
				Message:  fmt.Sprintf("必需的文档文件 '%s' 不存在", doc),
				Document: doc,
			})
		}
	}
}

func (c *ConsistencyChecker) featureExistsInCode(indicators []string) bool {
	for _, elem := range c.codeScanner.Elements() {
		for _, indicator := range indicators {
			if strings.Contains(elem.Name, indicator) || strings.Contains(elem.FilePath, indicator) {
				return true
			}
		}
	}
	return false
}

func (c *ConsistencyChecker) featureExistsInDocs(sections []string) bool {
	for _, section := range c.docParser.Sections() {
		for _, keyword := range sections {
			if strings.Contains(section.Title, keyword) || strings.Contains(section.Content, keyword) {
				return true
			}
		}
	}
	return false
}

func (c *ConsistencyChecker) extractReadmeVersion() string {
	path := filepath.Join(c.projectRoot, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	match := versionPattern.FindStringSubmatch(string(data))
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func (c *ConsistencyChecker) extractCodeVersion() string {
	path := filepath.Join(c.projectRoot, "frontend", "package.json")
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if match := pkgVersionPattern.FindStringSubmatch(line); len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func (c *ConsistencyChecker) buildResult() *models.CheckResult {
	bySeverity := make(map[string]int)
	byType := make(map[string]int)

	for _, issue := range c.issues {
		bySeverity[issue.Severity]++
		byType[issue.Type]++
	}

	score := c.calculateScore()

	return &models.CheckResult{
		TotalIssues:       len(c.issues),
		CodeElementsCount: len(c.codeScanner.Elements()),
		DocSectionsCount:  len(c.docParser.Sections()),
		Issues:            c.issues,
		IssuesBySeverity:  bySeverity,
		IssuesByType:      byType,
		ConsistencyScore:  score,
	}
}

func (c *ConsistencyChecker) calculateScore() float64 {
	importantNames := make(map[string]bool)
	for _, elem := range c.codeScanner.Elements() {
		if strings.HasSuffix(elem.Name, "Handler") || strings.HasSuffix(elem.Name, "Service") {
			importantNames[elem.Name] = false
		}
		if elem.Type == "component" {
			importantNames[elem.Name] = false
		}
	}

	if len(importantNames) == 0 {
		return 100
	}

	documented := 0
	for name := range importantNames {
		if c.isNameDocumented(name) {
			documented++
		}
	}

	coverage := float64(documented) / float64(len(importantNames))
	penalty := float64(len(c.issues)) * 0.05
	if penalty > 0.3 {
		penalty = 0.3
	}

	score := (coverage - penalty) * 100
	if score < 0 {
		score = 0
	}
	return score
}
