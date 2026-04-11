package downloadtemplate

import (
	"flag"
	"fmt"
	"github.com/franchb/htmgo/cli/htmgo/internal/dirutil"
	"github.com/franchb/htmgo/cli/htmgo/tasks/process"
	"github.com/franchb/htmgo/cli/htmgo/tasks/run"
	"github.com/franchb/htmgo/cli/htmgo/tasks/util"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func DownloadTemplate(outPath string) {
	cwd, _ := os.Getwd()

	flag.Parse()

	outPath = strings.ReplaceAll(outPath, "\n", "")
	outPath = strings.ReplaceAll(outPath, "\r", "")
	outPath = strings.ReplaceAll(outPath, " ", "-")
	outPath = strings.ToLower(outPath)

	if outPath == "" {
		fmt.Println("Please provide a name for your app.")
		return
	}

	templateName := "starter-template"
	templatePath := filepath.Join("templates", "starter")

	re := regexp.MustCompile(`[^a-zA-Z]+`)
	// Replace all non-alphabetic characters with an empty string
	newModuleName := re.ReplaceAllString(outPath, "")

	tempOut := newModuleName + "_temp_" + strconv.FormatInt(time.Now().Unix(), 10)

	fmt.Printf("Downloading template %s\n to %s", templateName, tempOut)

	err := process.Run(process.NewRawCommand("clone-template", "git clone https://github.com/franchb/htmgo --depth=1 "+tempOut, process.ExitOnError))

	if err != nil {
		log.Fatalf("Error cloning the template, error: %s\n", err.Error())
		return
	}

	slog.Debug("provided out path", slog.String("outPath", outPath))
	slog.Debug("new module name", slog.String("newModuleName", newModuleName))
	slog.Debug("cwd", slog.String("cwd", cwd))

	newDir := filepath.Join(cwd, outPath)

	slog.Debug("Copying template files to", slog.String("dir", newDir))

	dirutil.CopyDir(filepath.Join(tempOut, templatePath), newDir, func(path string, exists bool) bool {
		return true
	})

	dirutil.DeleteDir(tempOut)

	process.SetWorkingDir(newDir)

	slog.Debug("current working dir", slog.String("cwd", process.GetWorkingDir()))

	commands := [][]string{
		{"git", "init"},
	}

	for _, command := range commands {
		process.Run(process.NewRawCommand("", strings.Join(command, " "), process.ExitOnError))
	}

	_ = util.ReplaceTextInFile(filepath.Join(newDir, "go.mod"),
		fmt.Sprintf("module %s", templateName),
		fmt.Sprintf("module %s", newModuleName))

	removeReplaceDirectives(filepath.Join(newDir, "go.mod"))

	_ = util.ReplaceTextInDirRecursive(newDir, templateName, newModuleName, func(file string) bool {
		return strings.HasSuffix(file, ".go") || strings.HasPrefix(file, "Dockerfile")
	})

	fmt.Printf("Setting up the project in %s\n", newDir)
	process.SetWorkingDir(newDir)
	run.Setup()
	process.SetWorkingDir("")

	fmt.Println("Template downloaded successfully.")
	fmt.Println("To start the development server, run the following commands:")
	fmt.Printf("cd %s && htmgo watch\n", outPath)

	fmt.Printf("To build the project, run the following command:\n")
	fmt.Printf("cd %s && htmgo build\n", outPath)
}

// removeReplaceDirectives strips all "replace" lines from a go.mod file.
// Templates use replace directives for local development, but standalone
// projects must resolve dependencies from the module proxy.
func removeReplaceDirectives(goModPath string) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return
	}
	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "replace ") {
			kept = append(kept, line)
		}
	}
	os.WriteFile(goModPath, []byte(strings.Join(kept, "\n")), 0644)
}
