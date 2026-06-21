package github

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	PullRequestTemplate = "pull_request_template"
	IssueTemplate       = "issue_template"
	githubTemplateDir   = ".github"
	docsDir             = "docs"
)

func ReadTemplate(kind, workdir string) (body string, err error) {
	templateDir := filepath.Join(workdir, githubTemplateDir)

	path, err := getFilePath(templateDir, kind)
	if err != nil || path == "" {
		docsDir := filepath.Join(workdir, docsDir)
		path, err = getFilePath(docsDir, kind)
	}
	if err != nil || path == "" {
		path, err = getFilePath(workdir, kind)
	}

	if path != "" {
		body, err = readContentsFromFile(path)
	}
	return
}

func ListTemplates(kind, workdir string) ([]string, error) {
	var templates []string
	seen := make(map[string]bool)

	searchDirs := []string{
		filepath.Join(workdir, githubTemplateDir, kind+"s"),
		filepath.Join(workdir, githubTemplateDir),
		filepath.Join(workdir, docsDir),
		workdir,
	}

	for _, dir := range searchDirs {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			lowerName := strings.ToLower(name)
			if !strings.Contains(lowerName, strings.ToLower(kind)) {
				continue
			}
			if !strings.HasSuffix(lowerName, ".md") && !strings.HasSuffix(lowerName, ".txt") {
				continue
			}
			baseName := strings.TrimSuffix(name, ".md")
			baseName = strings.TrimSuffix(baseName, ".txt")
			if !seen[baseName] {
				seen[baseName] = true
				templates = append(templates, baseName)
			}
		}
	}

	sort.Strings(templates)
	return templates, nil
}

func ReadTemplateByName(kind, templateName, workdir string) (string, error) {
	searchDirs := []string{
		filepath.Join(workdir, githubTemplateDir, kind+"s"),
		filepath.Join(workdir, githubTemplateDir),
		filepath.Join(workdir, docsDir),
		workdir,
	}

	for _, dir := range searchDirs {
		for _, ext := range []string{".md", ".txt", ""} {
			path := filepath.Join(dir, templateName+ext)
			if _, err := os.Stat(path); err == nil {
				return readContentsFromFile(path)
			}
		}
	}

	return "", fmt.Errorf("template not found: %s", templateName)
}

func ReplaceTemplatePlaceholders(content string, placeholders map[string]string) string {
	for key, value := range placeholders {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}
	return content
}

type sortedFiles []os.FileInfo

func (s sortedFiles) Len() int {
	return len(s)
}
func (s sortedFiles) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortedFiles) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(s[i].Name()), strings.ToLower(s[j].Name())) > 0
}

func getFilePath(dir, pattern string) (found string, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	sort.Sort(sortedFiles(files))

	for _, file := range files {
		fileName := file.Name()
		path := strings.TrimSuffix(fileName, ".md")
		path = strings.TrimSuffix(path, ".txt")

		if strings.EqualFold(pattern, path) {
			found = filepath.Join(dir, fileName)
			return
		}
	}
	return
}

func readContentsFromFile(filename string) (contents string, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if strings.HasSuffix(err.Error(), " is a directory") {
			err = nil
		}
		return
	}

	contents = strings.Replace(string(content), "\r\n", "\n", -1)
	contents = strings.TrimSuffix(contents, "\n")
	return
}
