package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coder100001/etf-insight/tools/doccheck/models"
)

var (
	headerPattern = regexp.MustCompile(`^(#+)\s+(.+)$`)
)

var docFilePatterns = []string{
	"AGENTS.md",
	"agents.md",
	"README.md",
	"README_EN.md",
}

var docDirPatterns = []string{
	"README.md",
	"readme.md",
	"README.txt",
	"readme.txt",
}

var skipDocDirs = []string{
	"node_modules", "vendor", ".git", "__pycache__",
	"dist", "build", ".next", "coverage",
}

type DocumentParser struct {
	projectRoot string
	sections    []models.DocumentSection
}

func NewDocumentParser(projectRoot string) *DocumentParser {
	return &DocumentParser{
		projectRoot: projectRoot,
		sections:    make([]models.DocumentSection, 0),
	}
}

func (p *DocumentParser) Sections() []models.DocumentSection {
	return p.sections
}

func (p *DocumentParser) Parse() error {
	if err := p.parseRootDocs(); err != nil {
		return err
	}

	if err := p.parseSubDirDocs(); err != nil {
		return err
	}

	return nil
}

func (p *DocumentParser) parseRootDocs() error {
	for _, pattern := range docFilePatterns {
		path := filepath.Join(p.projectRoot, pattern)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if err := p.parseFile(path); err != nil {
				continue
			}
		}
	}
	return nil
}

func (p *DocumentParser) parseSubDirDocs() error {
	dirs := []string{"backend", "frontend", "docs"}
	for _, dir := range dirs {
		dirPath := filepath.Join(p.projectRoot, dir)
		if info, err := os.Stat(dirPath); err != nil || !info.IsDir() {
			continue
		}
		if err := p.walkDirForDocs(dirPath); err != nil {
			continue
		}
	}
	return nil
}

func (p *DocumentParser) walkDirForDocs(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			for _, skip := range skipDocDirs {
				if strings.Contains(path, skip) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		name := strings.ToLower(info.Name())
		if strings.HasPrefix(name, "readme") || name == "agents.md" {
			_ = p.parseFile(path)
		}

		return nil
	})
}

func (p *DocumentParser) parseFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	relPath, _ := filepath.Rel(p.projectRoot, path)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	sections := p.extractSections(lines, relPath)
	p.sections = append(p.sections, sections...)

	return nil
}

func (p *DocumentParser) extractSections(lines []string, filePath string) []models.DocumentSection {
	var sections []models.DocumentSection

	var current *models.DocumentSection
	var contentBuilder strings.Builder

	for i, line := range lines {
		if match := headerPattern.FindStringSubmatch(line); match != nil {
			if current != nil {
				current.Content = contentBuilder.String()
				current.LineEnd = i
				sections = append(sections, *current)
			}

			title := strings.TrimSpace(match[2])
			current = &models.DocumentSection{
				Title:     title,
				FilePath:  filePath,
				LineStart: i + 1,
			}
			contentBuilder.Reset()
		} else if current != nil {
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	if current != nil {
		current.Content = contentBuilder.String()
		current.LineEnd = len(lines)
		sections = append(sections, *current)
	}

	return sections
}