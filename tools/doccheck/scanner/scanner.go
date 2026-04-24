package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coder100001/etf-insight/tools/doccheck/models"
)

var (
	goFuncPattern      = regexp.MustCompile(`func\s+(?:\((\w+)\s+\*?(\w+)\)\s+)?(\w+)\s*\(`)
	goStructPattern    = regexp.MustCompile(`type\s+(\w+)\s+struct\s*\{`)
	goInterfacePattern = regexp.MustCompile(`type\s+(\w+)\s+interface\s*\{`)
	goCommentPattern   = regexp.MustCompile(`^\s*//\s*(.+)`)

	tsFuncCompPattern  = regexp.MustCompile(`const\s+(\w+)\s*[:=]\s*(?:React\.)?(?:memo\()?\s*\(?(?:function\s*\(|\(\))`)
	tsFuncPattern      = regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(`)
	tsInterfacePattern = regexp.MustCompile(`interface\s+(\w+)\s*\{`)

	skipDirPatterns = []string{
		"node_modules", "vendor", ".git", "__pycache__",
		"dist", "build", ".next", "coverage",
	}

	skipStructSuffixes = []string{
		"Request", "Response", "Config", "Input", "Output",
		"Result", "DTO", "VO",
	}

	coreStructPrefixes = []string{
		"ETF", "Portfolio", "Asset", "Backtest", "Factor",
		"Exchange", "Risk", "Optimization", "Holding",
	}
)

type CodeScanner struct {
	projectRoot string
	elements    []models.CodeElement
}

func NewCodeScanner(projectRoot string) *CodeScanner {
	return &CodeScanner{
		projectRoot: projectRoot,
		elements:    make([]models.CodeElement, 0),
	}
}

func (s *CodeScanner) Elements() []models.CodeElement {
	return s.elements
}

func (s *CodeScanner) Scan() error {
	backendPath := filepath.Join(s.projectRoot, "backend")
	if info, err := os.Stat(backendPath); err == nil && info.IsDir() {
		if err := s.scanGoDir(backendPath); err != nil {
			return fmt.Errorf("扫描Go代码失败: %w", err)
		}
	}

	frontendPath := filepath.Join(s.projectRoot, "frontend", "src")
	if info, err := os.Stat(frontendPath); err == nil && info.IsDir() {
		if err := s.scanTSDir(frontendPath); err != nil {
			return fmt.Errorf("扫描TypeScript代码失败: %w", err)
		}
	}

	return nil
}

func (s *CodeScanner) shouldSkip(path string) bool {
	for _, pattern := range skipDirPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func (s *CodeScanner) scanGoDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if s.shouldSkip(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return s.parseGoFile(path)
	})
}

func (s *CodeScanner) parseGoFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	relPath, _ := filepath.Rel(s.projectRoot, path)
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	content := strings.Join(lines, "\n")

	s.extractGoFunctions(relPath, content, lines)
	s.extractGoStructs(relPath, content, lines)
	s.extractGoInterfaces(relPath, content, lines)

	return nil
}

func (s *CodeScanner) extractGoFunctions(relPath, content string, lines []string) {
	matches := goFuncPattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		var receiverType, name string
		if match[4] != -1 && match[5] != -1 {
			receiverType = content[match[4]:match[5]]
		}
		if match[6] != -1 && match[7] != -1 {
			name = content[match[6]:match[7]]
		}

		if !s.isExported(name) {
			continue
		}
		if strings.HasPrefix(name, "New") && !s.isCoreStruct(name[3:]) {
			continue
		}
		if strings.HasPrefix(name, "test") || strings.HasPrefix(name, "Test") {
			continue
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1
		comment := s.findGoComment(lines, lineNum)

		elementType := "function"
		if strings.HasSuffix(receiverType, "Handler") {
			elementType = "handler"
		} else if strings.HasSuffix(receiverType, "Service") {
			elementType = "service"
		}

		s.elements = append(s.elements, models.CodeElement{
			Name:        name,
			Type:        elementType,
			FilePath:    relPath,
			LineStart:   lineNum,
			Description: comment,
		})
	}
}

func (s *CodeScanner) isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func (s *CodeScanner) extractGoStructs(relPath, content string, lines []string) {
	matches := goStructPattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]

		if s.isSkipStruct(name) && !s.isCoreStruct(name) {
			continue
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1
		comment := s.findGoComment(lines, lineNum)

		elementType := "model"
		if strings.HasSuffix(name, "Handler") && strings.Contains(relPath, "handlers/") {
			elementType = "handler"
		} else if strings.HasSuffix(name, "Service") && strings.Contains(relPath, "services/") {
			elementType = "service"
		}

		s.elements = append(s.elements, models.CodeElement{
			Name:        name,
			Type:        elementType,
			FilePath:    relPath,
			LineStart:   lineNum,
			Description: comment,
		})
	}
}

func (s *CodeScanner) isSkipStruct(name string) bool {
	for _, suffix := range skipStructSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func (s *CodeScanner) isCoreStruct(name string) bool {
	for _, prefix := range coreStructPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (s *CodeScanner) extractGoInterfaces(relPath, content string, lines []string) {
	matches := goInterfacePattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1
		comment := s.findGoComment(lines, lineNum)

		s.elements = append(s.elements, models.CodeElement{
			Name:        name,
			Type:        "interface",
			FilePath:    relPath,
			LineStart:   lineNum,
			Description: comment,
		})
	}
}

func (s *CodeScanner) findGoComment(lines []string, lineNum int) string {
	comment := ""
	for i := lineNum - 2; i >= 0 && i >= lineNum-10; i-- {
		line := strings.TrimSpace(lines[i])
		if goCommentPattern.MatchString(line) {
			submatch := goCommentPattern.FindStringSubmatch(line)
			comment = submatch[1] + " " + comment
		} else {
			break
		}
	}
	return strings.TrimSpace(comment)
}

func (s *CodeScanner) scanTSDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if s.shouldSkip(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
			return nil
		}

		base := strings.ToLower(info.Name())
		if strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
			return nil
		}

		return s.parseTSFile(path)
	})
}

func (s *CodeScanner) parseTSFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	relPath, _ := filepath.Rel(s.projectRoot, path)
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	content := strings.Join(lines, "\n")

	s.extractTSComponents(relPath, content, lines)
	s.extractTSFunctions(relPath, content, lines)
	s.extractTSInterfaces(relPath, content, lines)

	return nil
}

func (s *CodeScanner) extractTSComponents(relPath, content string, lines []string) {
	matches := tsFuncCompPattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		s.elements = append(s.elements, models.CodeElement{
			Name:      name,
			Type:      "component",
			FilePath:  relPath,
			LineStart: lineNum,
		})
	}
}

func (s *CodeScanner) extractTSFunctions(relPath, content string, lines []string) {
	matches := tsFuncPattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		if strings.HasPrefix(name, "_") {
			continue
		}
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		elementType := "function"
		if strings.Contains(relPath, "services/") {
			elementType = "service"
		}

		s.elements = append(s.elements, models.CodeElement{
			Name:      name,
			Type:      elementType,
			FilePath:  relPath,
			LineStart: lineNum,
		})
	}
}

func (s *CodeScanner) extractTSInterfaces(relPath, content string, lines []string) {
	matches := tsInterfacePattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		s.elements = append(s.elements, models.CodeElement{
			Name:      name,
			Type:      "interface",
			FilePath:  relPath,
			LineStart: lineNum,
		})
	}
}
