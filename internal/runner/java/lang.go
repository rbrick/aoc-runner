package java

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const (
	DefaultSolutionCode = `class Solution {
   public String partOne(Reader reader) {
       return "";
   }

   public String partTwo(Reader reader) {
       return "";
   }
}`
)

//go:embed java.template
var rawJavaTemplate string

// parse the java template
var JavaTemplate *template.Template = template.Must(template.New("java").Parse(rawJavaTemplate))

type TemplateData struct {
	Solution string
}

type JavaLanguage struct{}

// So in this context
func (*JavaLanguage) Build(ctx context.Context, code string) (string, error) {
	// first build and check for errors
	tempFile, err := os.CreateTemp(os.TempDir(), "code.java")

	if err != nil {
		return "", err
	}

	path := filepath.Join(os.TempDir(), "code.java")

	JavaTemplate.ExecuteTemplate(tempFile, "java", &TemplateData{
		Solution: code,
	})

	cmd := exec.CommandContext(ctx, "javac", path)
	cmd.Run()
	return "", nil
}

func (*JavaLanguage) Run(ctx context.Context, code string) (string, error)
