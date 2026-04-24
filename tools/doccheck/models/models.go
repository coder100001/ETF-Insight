package models

type CodeElement struct {
	Name        string
	Type        string
	FilePath    string
	LineStart   int
	LineEnd     int
	Description string
}

type DocumentSection struct {
	Title     string
	Content   string
	FilePath  string
	LineStart int
	LineEnd   int
}

type Issue struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Message        string `json:"message"`
	Document       string `json:"document,omitempty"`
	DocumentLine   int    `json:"document_line,omitempty"`
	ElementName    string `json:"element_name,omitempty"`
	ElementType    string `json:"element_type,omitempty"`
	ElementFile    string `json:"element_file,omitempty"`
	Feature        string `json:"feature,omitempty"`
	ReadmeVersion  string `json:"readme_version,omitempty"`
	CodeVersion    string `json:"code_version,omitempty"`
}

type CheckResult struct {
	TotalIssues        int              `json:"total_issues"`
	CodeElementsCount  int              `json:"code_elements_count"`
	DocSectionsCount   int              `json:"document_sections_count"`
	Issues             []Issue          `json:"issues"`
	IssuesBySeverity   map[string]int   `json:"issues_by_severity"`
	IssuesByType       map[string]int   `json:"issues_by_type"`
	ConsistencyScore   float64          `json:"consistency_score"`
}

type MappingRule struct {
	Description           string   `json:"description"`
	DocumentSection       string   `json:"document_section"`
	RequiredDocumentation []string `json:"required_documentation"`
}

type SpecificMapping struct {
	DocumentFiles []string `json:"document_files"`
	Sections      []string `json:"sections"`
	Description   string   `json:"description"`
}

type FeatureMapping struct {
	CodeIndicators   []string `json:"code_indicators"`
	DocumentSections []string `json:"document_sections"`
	Required         bool     `json:"required"`
}

type MappingConfig struct {
	Version         string                    `json:"version"`
	MappingRules    map[string]map[string]MappingRule `json:"mapping_rules"`
	FeatureMappings map[string]FeatureMapping `json:"feature_mappings"`
	ValidationRules struct {
		RequiredDocuments   []string `json:"required_documents"`
		MinCoverageThreshold float64 `json:"min_coverage_threshold"`
		HighSeverityIssues  []string `json:"high_severity_issues"`
	} `json:"validation_rules"`
}