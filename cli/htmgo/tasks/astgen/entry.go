package astgen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/franchb/htmgo/cli/htmgo/internal/dirutil"
	"github.com/franchb/htmgo/cli/htmgo/tasks/process"
	"github.com/franchb/htmgo/framework/v2/h"
	"golang.org/x/mod/modfile"
)

type Page struct {
	Path     string
	FuncName string
	Package  string
	Import   string
}

type Partial struct {
	FuncName string
	Package  string
	Import   string
	Path     string
}

const GeneratedDirName = "__htmgo"
const FiberModuleName = "github.com/gofiber/fiber/v3"
const ModuleName = "github.com/franchb/htmgo/framework/v2/h"

var PackageName = fmt.Sprintf("package %s", GeneratedDirName)
var GeneratedFileLine = fmt.Sprintf("// Package %s THIS FILE IS GENERATED. DO NOT EDIT.", GeneratedDirName)

func toPascaleCase(input string) string {
	words := strings.Split(input, "_")
	for i := range words {
		words[i] = strings.Title(strings.ToLower(words[i]))
	}
	return strings.Join(words, "")
}

func isValidGoVariableName(name string) bool {
	// Variable name must not be empty
	if name == "" {
		return false
	}
	// First character must be a letter or underscore
	if !unicode.IsLetter(rune(name[0])) && name[0] != '_' {
		return false
	}
	// Remaining characters must be letters, digits, or underscores
	for _, char := range name[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}
	return true
}

func normalizePath(path string) string {
	return strings.ReplaceAll(path, `\`, "/")
}

func sliceCommonPrefix(dir1, dir2 string) string {
	// Use filepath.Clean to normalize the paths
	dir1 = filepath.Clean(dir1)
	dir2 = filepath.Clean(dir2)

	// Find the common prefix
	commonPrefix := dir1
	if len(dir1) > len(dir2) {
		commonPrefix = dir2
	}

	for !strings.HasPrefix(dir1, commonPrefix) {
		commonPrefix = filepath.Dir(commonPrefix)
	}

	// Slice off the common prefix
	slicedDir1 := strings.TrimPrefix(dir1, commonPrefix)
	slicedDir2 := strings.TrimPrefix(dir2, commonPrefix)

	// Remove leading slashes
	slicedDir1 = strings.TrimPrefix(slicedDir1, string(filepath.Separator))
	slicedDir2 = strings.TrimPrefix(slicedDir2, string(filepath.Separator))

	// Return the longer one
	if len(slicedDir1) > len(slicedDir2) {
		return normalizePath(slicedDir1)
	}
	return normalizePath(slicedDir2)
}

func hasOnlyReqContextParam(funcType *ast.FuncType) bool {
	if len(funcType.Params.List) != 1 {
		return false
	}
	if funcType.Params.List[0].Names == nil {
		return false
	}
	if len(funcType.Params.List[0].Names) != 1 {
		return false
	}
	t := funcType.Params.List[0].Type
	name, ok := t.(*ast.StarExpr)
	if !ok {
		return false
	}
	selectorExpr, ok := name.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "h" && selectorExpr.Sel.Name == "RequestContext"
}

// findPagesAndPartials walks the directory tree once, parsing each Go file a single time,
// and classifies exported functions into pages (returning *h.Page) and partials (returning *h.Partial).
// Files under pagesDir are checked for both pages and partials; all other files only for partials.
func findPagesAndPartials(rootDir string, pagesDir string, partialPredicate func(partial Partial) bool) ([]Page, []Partial, error) {
	var pages []Page
	var partials []Partial
	cwd := process.GetWorkingDir()

	// Make pagesDir absolute so it matches the absolute paths from filepath.Walk.
	absPagesDir := pagesDir
	if !filepath.IsAbs(pagesDir) {
		absPagesDir = filepath.Join(rootDir, pagesDir)
	}
	normalizedPagesDir := normalizePath(absPagesDir)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return err
		}

		// Determine if this file is under the pages directory.
		normalizedPath := normalizePath(path)
		isUnderPages := strings.HasPrefix(normalizedPath, normalizedPagesDir+"/") ||
			normalizedPath == normalizedPagesDir

		ast.Inspect(node, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if !funcDecl.Name.IsExported() {
				return true
			}
			if funcDecl.Type.Results == nil {
				return true
			}
			if !hasOnlyReqContextParam(funcDecl.Type) {
				return true
			}

			for _, result := range funcDecl.Type.Results.List {
				starExpr, ok := result.Type.(*ast.StarExpr)
				if !ok {
					continue
				}
				selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				ident, ok := selectorExpr.X.(*ast.Ident)
				if !ok || ident.Name != "h" {
					continue
				}

				switch selectorExpr.Sel.Name {
				case "Page":
					if isUnderPages {
						// Use paths relative to cwd to match the original behavior
						// of findPublicFuncsReturningHPage which walked from "pages/".
						relPath, _ := filepath.Rel(cwd, path)
						pages = append(pages, Page{
							Package:  node.Name.Name,
							Import:   normalizePath(filepath.Dir(relPath)),
							Path:     normalizePath(relPath),
							FuncName: funcDecl.Name.Name,
						})
					}
				case "Partial":
					p := Partial{
						Package:  node.Name.Name,
						Path:     normalizePath(sliceCommonPrefix(cwd, path)),
						Import:   sliceCommonPrefix(cwd, normalizePath(filepath.Dir(path))),
						FuncName: funcDecl.Name.Name,
					}
					if partialPredicate(p) {
						partials = append(partials, p)
					}
				}
				break
			}
			return true
		})

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return pages, partials, nil
}

func buildGetPartialFromContext(builder *CodeBuilder, partials []Partial) {
	moduleName := GetModuleName()

	var routerHandlerMethod = func(path string, caller string) string {
		return fmt.Sprintf(`
			router.All("%s", func(c fiber.Ctx) error {
				cc := h.GetRequestContext(c)
				partial := %s(cc)
				if partial == nil {
					return c.SendStatus(404)
				}
				return h.PartialView(c, partial)
			})`, path, caller)
	}

	handlerMethods := make([]string, 0)

	for _, f := range partials {
		caller := fmt.Sprintf("%s.%s", f.Package, f.FuncName)
		path := fmt.Sprintf("/%s/%s.%s", moduleName, f.Import, f.FuncName)
		handlerMethods = append(handlerMethods, routerHandlerMethod(path, caller))
	}

	registerFunction := fmt.Sprintf(`
		func RegisterPartials(router *fiber.App) {
				%s
		}
	`, strings.Join(handlerMethods, "\n"))

	builder.AppendLine(registerFunction)
}

func writeGenerated() error {
	config := dirutil.GetConfig()

	cwd := process.GetWorkingDir()
	pages, partials, err := findPagesAndPartials(cwd, "pages", func(partial Partial) bool {
		return partial.FuncName != "GetPartialFromContext"
	})

	if err != nil {
		return err
	}

	partials = h.Filter(partials, func(partial Partial) bool {
		return !dirutil.IsGlobExclude(partial.Path, config.AutomaticPartialRoutingIgnore)
	})

	pages = h.Filter(pages, func(page Page) bool {
		return !dirutil.IsGlobExclude(page.Path, config.AutomaticPageRoutingIgnore)
	})

	writePartialsFile(partials)
	writePagesFile(pages)
	return nil
}

func writePartialsFile(partials []Partial) {
	builder := NewCodeBuilder(nil)
	builder.AppendLine(GeneratedFileLine)
	builder.AppendLine(PackageName)
	builder.AddImport(FiberModuleName)

	if len(partials) > 0 {
		builder.AddImport(ModuleName)
	}

	moduleName := GetModuleName()
	for _, partial := range partials {
		builder.AddImport(fmt.Sprintf(`%s/%s`, moduleName, partial.Import))
	}

	buildGetPartialFromContext(builder, partials)

	WriteFile(filepath.Join(GeneratedDirName, "partials-generated.go"), func() string {
		return builder.String()
	})
}

func formatRoute(path string) string {
	path = strings.TrimSuffix(path, "index.go")
	path = strings.TrimSuffix(path, ".go")
	path = strings.TrimPrefix(path, "pages/")
	path = strings.TrimPrefix(path, "pages\\")
	path = strings.ReplaceAll(path, "$", ":")
	path = strings.ReplaceAll(path, "_", "/")
	path = strings.ReplaceAll(path, ".", "/")
	path = strings.ReplaceAll(path, "\\", "/")

	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = filepath.Join("/", path)
	}
	if strings.HasSuffix(path, "/") {
		return strings.ReplaceAll(path[:len(path)-1], `\`, "/")
	}
	return strings.ReplaceAll(filepath.Clean(path), `\`, "/")
}

func writePagesFile(pages []Page) {
	builder := NewCodeBuilder(nil)
	builder.AppendLine(GeneratedFileLine)
	builder.AppendLine(PackageName)
	builder.AddImport(FiberModuleName)

	if len(pages) > 0 {
		builder.AddImport(ModuleName)
	}

	for _, page := range pages {
		if page.Import != "" {
			moduleName := GetModuleName()
			builder.AddImport(
				fmt.Sprintf(`%s/%s`, moduleName, page.Import),
			)
		}
	}

	fName := "RegisterPages"
	body := `
	`

	for _, page := range pages {
		call := fmt.Sprintf("%s.%s", page.Package, page.FuncName)

		body += fmt.Sprintf(
			`
			router.Get("%s", func(c fiber.Ctx) error {
				cc := h.GetRequestContext(c)
				return h.HtmlView(c, %s(cc))
			})
			`, formatRoute(page.Path), call,
		)
	}

	f := Function{
		Name: fName,
		Parameters: []NameType{
			{Name: "router", Type: "*fiber.App"},
		},
		Body: body,
	}

	builder.Append(builder.BuildFunction(f))

	WriteFile(filepath.Join(GeneratedDirName, "pages-generated.go"), func() string {
		return builder.String()
	})
}

func writeAssetsFile() {
	cwd := process.GetWorkingDir()
	config := dirutil.GetConfig()

	slog.Debug("writing assets file", slog.String("cwd", cwd), slog.String("config", config.PublicAssetPath))

	distAssets := filepath.Join(cwd, "assets", "dist")
	hasAssets := false

	builder := strings.Builder{}

	builder.WriteString(`package assets`)
	builder.WriteString("\n")

	filepath.WalkDir(distAssets, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		path = strings.ReplaceAll(path, distAssets, "")
		httpUrl := normalizePath(fmt.Sprintf("%s%s", config.PublicAssetPath, path))

		path = normalizePath(path)
		path = strings.ReplaceAll(path, "/", "_")
		path = strings.ReplaceAll(path, "//", "_")

		name := strings.ReplaceAll(path, ".", "_")
		name = strings.ReplaceAll(name, "-", "_")

		name = toPascaleCase(name)

		if isValidGoVariableName(name) {
			builder.WriteString(fmt.Sprintf(`const %s = "%s"`, name, httpUrl))
			builder.WriteString("\n")
			hasAssets = true
		}

		return nil
	})

	builder.WriteString("\n")

	str := builder.String()

	if hasAssets {
		WriteFile(filepath.Join(GeneratedDirName, "assets", "assets-generated.go"), func() string {
			return str
		})
	}

}

func HasModuleFile(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func CheckPagesDirectory(path string) error {
	pagesPath := filepath.Join(path, "pages")
	_, err := os.Stat(pagesPath)
	if err != nil {
		return fmt.Errorf("The directory pages does not exist.")
	}

	return nil
}

var (
	cachedModuleName string
	moduleNameMu     sync.Mutex
)

// GetModuleName returns the module name from go.mod, caching the result on
// success. Failures are NOT cached so that watch-mode retries succeed once
// the project is fully initialized.
func GetModuleName() string {
	moduleNameMu.Lock()
	defer moduleNameMu.Unlock()

	if cachedModuleName != "" {
		return cachedModuleName
	}

	wd := process.GetWorkingDir()
	modPath := filepath.Join(wd, "go.mod")

	if HasModuleFile(modPath) == false {
		fmt.Fprintf(os.Stderr, "Module not found: go.mod file does not exist.")
		return ""
	}

	checkDir := CheckPagesDirectory(wd)
	if checkDir != nil {
		fmt.Fprintf(os.Stderr, "%s", checkDir.Error())
		return ""
	}

	goModBytes, err := os.ReadFile(modPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading go.mod: %v\n", err)
		return ""
	}
	cachedModuleName = modfile.ModulePath(goModBytes)
	return cachedModuleName
}

func GenAst(flags ...process.RunFlag) error {
	if GetModuleName() == "" {
		if slices.Contains(flags, process.ExitOnError) {
			os.Exit(1)
		}
		return fmt.Errorf("error getting module name")
	}
	if err := writeGenerated(); err != nil {
		return err
	}
	writeAssetsFile()

	WriteFile("__htmgo/setup-generated.go", func() string {

		return fmt.Sprintf(`
			// Package __htmgo THIS FILE IS GENERATED. DO NOT EDIT.
			package __htmgo

			import (
				"%s"
			)

			func Register(r *fiber.App) {
				RegisterPartials(r)
				RegisterPages(r)
			}
		`, FiberModuleName)
	})

	return nil
}
