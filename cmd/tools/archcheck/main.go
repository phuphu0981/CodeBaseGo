package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

type Severity string

const (
	SevError   Severity = "ERROR"
	SevWarning Severity = "WARN"
)

type Violation struct {
	Severity Severity
	Rule     string
	File     string
	Line     int
	Message  string
}

type Checker struct {
	rootDir    string
	strict     bool
	verbose    bool
	violations []Violation
	fset       *token.FileSet
	passCount  int
}

func NewChecker(rootDir string, strict, verbose bool) *Checker {
	return &Checker{
		rootDir: rootDir,
		strict:  strict,
		verbose: verbose,
		fset:    token.NewFileSet(),
	}
}

func (c *Checker) AddViolation(sev Severity, rule, file string, pos token.Pos, msg string) {
	line := 0
	if pos.IsValid() {
		line = c.fset.Position(pos).Line
	}
	c.violations = append(c.violations, Violation{
		Severity: sev,
		Rule:     rule,
		File:     file,
		Line:     line,
		Message:  msg,
	})
}

func (c *Checker) Pass(rule, desc string) {
	c.passCount++
	if c.verbose {
		fmt.Printf("  %s[PASS]%s %s: %s\n", colorGreen, colorReset, rule, desc)
	}
}

func isGeneratedFile(path string) bool {
	if strings.HasSuffix(path, "_gen.go") || strings.HasSuffix(path, "exec.go") {
		return true
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i := 0; i < 5 && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "Code generated") && strings.Contains(line, "DO NOT EDIT") {
			return true
		}
	}
	return false
}

func main() {
	strictFlag := flag.Bool("strict", false, "Fail on warnings as well as errors")
	verboseFlag := flag.Bool("verbose", false, "Show detailed passing checks")
	dirFlag := flag.String("dir", ".", "Root directory of the project to check")
	flag.Parse()

	checker := NewChecker(*dirFlag, *strictFlag, *verboseFlag)
	checker.RunAll()
}

func (c *Checker) RunAll() {
	fmt.Printf("%s%s=================================================================%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s    CodebaseGo Architecture & Coding Standard Checker (archcheck)  %s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s=================================================================%s\n\n", colorBold, colorCyan, colorReset)

	c.checkPackageAndFileNaming()
	c.checkModuleStructures()
	c.checkLayerIsolationAndImports()
	c.checkServiceLayerRules()
	c.checkHandlerLayerRules()
	c.checkDTOSecurity()
	c.checkErrorHandlingAndPanics()
	c.checkRegisterAndDIContracts()

	c.printSummary()
}

// 1. Check Package & File Naming Conventions
func (c *Checker) checkPackageAndFileNaming() {
	fmt.Printf("%s[1/8] Checking Package & File Naming Conventions...%s\n", colorBold, colorReset)

	fileNameRegex := regexp.MustCompile(`^[a-z0-9_]+(\.[a-z0-9_]+)*\.go$`)
	pkgNameRegex := regexp.MustCompile(`^[a-z0-9]+(_test)?$`)

	err := filepath.Walk(c.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "vendor" || name == "bin" || name == "tmp" || name == "docs" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		relPath, _ := filepath.Rel(c.rootDir, path)
		baseName := filepath.Base(path)

		// Check filename
		if !fileNameRegex.MatchString(baseName) {
			c.AddViolation(SevError, "Naming.File", relPath, token.NoPos,
				fmt.Sprintf("File name '%s' should be lowercase snake_case (e.g. gorm_repository.go)", baseName))
		}

		// Parse AST for package name
		node, parseErr := parser.ParseFile(c.fset, path, nil, parser.PackageClauseOnly)
		if parseErr == nil && node != nil {
			pkgName := node.Name.Name
			if pkgName != "main" && !pkgNameRegex.MatchString(pkgName) {
				c.AddViolation(SevError, "Naming.Package", relPath, node.Package,
					fmt.Sprintf("Package name '%s' should be lowercase without underscores or camelCase (except standard '_test' suffix)", pkgName))
			}
		}

		return nil
	})

	if err != nil {
		c.AddViolation(SevError, "System.Walk", "", token.NoPos, err.Error())
	} else {
		c.Pass("Naming.Convention", "File and package naming conventions verified")
	}
}

// 2. Check Module 4-Layer Structures
func (c *Checker) checkModuleStructures() {
	fmt.Printf("%s[2/8] Checking Feature Module Structures...%s\n", colorBold, colorReset)

	modulesDir := filepath.Join(c.rootDir, "internal", "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		c.AddViolation(SevError, "Structure.ModulesDir", "internal/modules", token.NoPos, "Cannot read internal/modules directory")
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modName := entry.Name()
		modPath := filepath.Join(modulesDir, modName)
		relModPath := filepath.Join("internal", "modules", modName)

		// Check register.go (Mandatory for every module)
		registerFile := filepath.Join(modPath, "register.go")
		if _, err := os.Stat(registerFile); os.IsNotExist(err) {
			c.AddViolation(SevError, "Structure.RegisterRequired", filepath.Join(relModPath, "register.go"), token.NoPos,
				fmt.Sprintf("Module '%s' must contain a 'register.go' file", modName))
		}

		// Specialized modules can have custom structure (e.g. health, graphql)
		if modName == "health" || modName == "graphql" {
			c.Pass("Structure.SpecializedModule", fmt.Sprintf("Module '%s' has specialized structure", modName))
			continue
		}

		// Core files required for all standard HTTP feature modules
		coreFiles := []string{"dto.go", "service.go", "handler.go"}
		for _, file := range coreFiles {
			target := filepath.Join(modPath, file)
			if _, err := os.Stat(target); os.IsNotExist(err) {
				c.AddViolation(SevWarning, "Structure.CoreFileMissing", filepath.Join(relModPath, file), token.NoPos,
					fmt.Sprintf("Feature module '%s' is missing recommended '%s'", modName, file))
			}
		}

		// Check if module is DB-backed (has entity.go or repository.go)
		entityFile := filepath.Join(modPath, "entity.go")
		repoFile := filepath.Join(modPath, "repository.go")
		hasDB := false
		if _, err := os.Stat(entityFile); err == nil {
			hasDB = true
		} else if _, err := os.Stat(repoFile); err == nil {
			hasDB = true
		}

		if hasDB {
			dbFiles := []string{"entity.go", "repository.go", "gorm_repository.go"}
			for _, file := range dbFiles {
				target := filepath.Join(modPath, file)
				if _, err := os.Stat(target); os.IsNotExist(err) {
					c.AddViolation(SevWarning, "Structure.DBFileMissing", filepath.Join(relModPath, file), token.NoPos,
						fmt.Sprintf("DB-backed module '%s' is missing '%s'", modName, file))
				}
			}
			c.Pass("Structure.DBModuleFiles", fmt.Sprintf("DB-backed module '%s' structure verified", modName))
		} else {
			c.Pass("Structure.StatelessModuleFiles", fmt.Sprintf("DB-less / Stateless module '%s' structure verified", modName))
		}
	}
}

// 3. Check Layer Isolation & Imports
func (c *Checker) checkLayerIsolationAndImports() {
	fmt.Printf("%s[3/8] Checking Layer Isolation & Import Boundaries...%s\n", colorBold, colorReset)

	_ = filepath.Walk(c.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || isGeneratedFile(path) {
			return nil
		}

		relPath, _ := filepath.Rel(c.rootDir, path)
		node, parseErr := parser.ParseFile(c.fset, path, nil, parser.ImportsOnly)
		if parseErr != nil || node == nil {
			return nil
		}

		// Check: internal/common must not import internal/modules or cmd
		if strings.HasPrefix(relPath, filepath.Join("internal", "common")) {
			for _, imp := range node.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if strings.Contains(importPath, "codebasego/internal/modules") || strings.Contains(importPath, "codebasego/cmd") {
					c.AddViolation(SevError, "Architecture.SharedKernelIsolation", relPath, imp.Pos(),
						fmt.Sprintf("internal/common must not import module/cmd packages: '%s'", importPath))
				}
			}
		}

		// Check: internal/platform must not import internal/modules or cmd
		if strings.HasPrefix(relPath, filepath.Join("internal", "platform")) {
			for _, imp := range node.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if strings.Contains(importPath, "codebasego/internal/modules") || strings.Contains(importPath, "codebasego/cmd") {
					c.AddViolation(SevError, "Architecture.PlatformIsolation", relPath, imp.Pos(),
						fmt.Sprintf("internal/platform must not import module/cmd packages: '%s'", importPath))
				}
			}
		}

		// Check: service.go must not import gin or gorm
		if strings.HasSuffix(relPath, "service.go") && strings.HasPrefix(relPath, filepath.Join("internal", "modules")) {
			for _, imp := range node.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if importPath == "github.com/gin-gonic/gin" {
					c.AddViolation(SevError, "Architecture.ServiceNoGin", relPath, imp.Pos(),
						"Service layer MUST NOT import 'github.com/gin-gonic/gin'")
				}
				if importPath == "gorm.io/gorm" {
					c.AddViolation(SevError, "Architecture.ServiceNoGorm", relPath, imp.Pos(),
						"Service layer MUST NOT import 'gorm.io/gorm' directly (depend on Repository interface)")
				}
			}
		}

		// Check: handler.go in standard feature modules must not import gorm or database/sql (health check module is exempt)
		if strings.HasSuffix(relPath, "handler.go") &&
			strings.HasPrefix(relPath, filepath.Join("internal", "modules")) &&
			!strings.Contains(relPath, filepath.Join("internal", "modules", "health")) {
			for _, imp := range node.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if importPath == "gorm.io/gorm" || importPath == "database/sql" {
					c.AddViolation(SevError, "Architecture.HandlerNoDB", relPath, imp.Pos(),
						"Feature Handler layer MUST NOT import database engines directly (no SQL in handler)")
				}
			}
		}

		return nil
	})
	c.Pass("Architecture.ImportBoundaries", "All package import boundaries verified")
}

// 4. Check Service Layer Rules
func (c *Checker) checkServiceLayerRules() {
	fmt.Printf("%s[4/8] Checking Service Layer Clean Architecture Rules...%s\n", colorBold, colorReset)

	_ = filepath.Walk(filepath.Join(c.rootDir, "internal", "modules"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "service.go") || isGeneratedFile(path) {
			return nil
		}

		relPath, _ := filepath.Rel(c.rootDir, path)
		node, parseErr := parser.ParseFile(c.fset, path, nil, parser.ParseComments)
		if parseErr != nil || node == nil {
			return nil
		}

		// Inspect Structs and Functions
		ast.Inspect(node, func(n ast.Node) bool {
			// Check Service struct fields (must not hold *gorm.DB)
			if typeSpec, ok := n.(*ast.TypeSpec); ok {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok && typeSpec.Name.Name == "Service" {
					for _, field := range structType.Fields.List {
						fieldTypeStr := fmt.Sprintf("%v", field.Type)
						if strings.Contains(fieldTypeStr, "gorm.DB") || strings.Contains(fieldTypeStr, "GormRepository") {
							c.AddViolation(SevError, "Architecture.ServiceDependency", relPath, field.Pos(),
								"Service struct must depend on Repository interface, not concrete GORM types")
						}
					}
				}
			}

			// Check Service methods (I/O and business methods should take context.Context as first param)
			if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Recv != nil && funcDecl.Name.IsExported() {
				recvName := ""
				for _, field := range funcDecl.Recv.List {
					if star, ok := field.Type.(*ast.StarExpr); ok {
						if ident, ok := star.X.(*ast.Ident); ok {
							recvName = ident.Name
						}
					}
				}

				// If receiver is *Service and method is not pure token validation / utility
				if recvName == "Service" && funcDecl.Name.Name != "ValidateToken" {
					params := funcDecl.Type.Params.List
					if len(params) == 0 {
						c.AddViolation(SevWarning, "Architecture.ServiceContext", relPath, funcDecl.Pos(),
							fmt.Sprintf("Service method '%s' should receive 'ctx context.Context' as first parameter", funcDecl.Name.Name))
					} else {
						firstParam := params[0]
						paramTypeStr := ""
						if sel, ok := firstParam.Type.(*ast.SelectorExpr); ok {
							paramTypeStr = fmt.Sprintf("%v.%v", sel.X, sel.Sel)
						} else if ident, ok := firstParam.Type.(*ast.Ident); ok {
							paramTypeStr = ident.Name
						}
						if paramTypeStr != "context.Context" {
							c.AddViolation(SevWarning, "Architecture.ServiceContext", relPath, firstParam.Pos(),
								fmt.Sprintf("Service method '%s' first parameter should be 'context.Context', got '%s'", funcDecl.Name.Name, paramTypeStr))
						}
					}
				}
			}

			return true
		})

		return nil
	})
	c.Pass("Architecture.ServiceRules", "Service layer Clean Architecture compliance verified")
}

// 5. Check Handler Layer Rules (Swagger Annotations & No SQL)
func (c *Checker) checkHandlerLayerRules() {
	fmt.Printf("%s[5/8] Checking Handler Layer & Swagger Documentation...%s\n", colorBold, colorReset)

	_ = filepath.Walk(filepath.Join(c.rootDir, "internal", "modules"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "handler.go") || isGeneratedFile(path) {
			return nil
		}

		relPath, _ := filepath.Rel(c.rootDir, path)
		node, parseErr := parser.ParseFile(c.fset, path, nil, parser.ParseComments)
		if parseErr != nil || node == nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Recv != nil && funcDecl.Name.IsExported() {
				recvName := ""
				for _, field := range funcDecl.Recv.List {
					if star, ok := field.Type.(*ast.StarExpr); ok {
						if ident, ok := star.X.(*ast.Ident); ok {
							recvName = ident.Name
						}
					}
				}

				if recvName == "Handler" {
					doc := funcDecl.Doc.Text()
					hasRouter := strings.Contains(doc, "@Router") || strings.Contains(doc, "@Summary")
					if !hasRouter {
						c.AddViolation(SevWarning, "Documentation.SwaggerAnnotation", relPath, funcDecl.Pos(),
							fmt.Sprintf("Public Handler method '%s' should have Swagger annotations (@Summary, @Router)", funcDecl.Name.Name))
					}
				}
			}
			return true
		})

		return nil
	})
	c.Pass("Documentation.Swagger", "Handler Swagger documentation annotations verified")
}

// 6. Check DTO Security (No Passwords or Secrets in Response DTOs)
func (c *Checker) checkDTOSecurity() {
	fmt.Printf("%s[6/8] Checking DTO Security (Sensitive Fields in Response)...%s\n", colorBold, colorReset)

	sensitiveRegex := regexp.MustCompile(`(?i)^(password|passwordhash|hashedpassword|secret|secretkey)$`)

	_ = filepath.Walk(filepath.Join(c.rootDir, "internal", "modules"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "dto.go") || isGeneratedFile(path) {
			return nil
		}

		relPath, _ := filepath.Rel(c.rootDir, path)
		node, parseErr := parser.ParseFile(c.fset, path, nil, 0)
		if parseErr != nil || node == nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			if typeSpec, ok := n.(*ast.TypeSpec); ok {
				typeName := typeSpec.Name.Name
				if strings.Contains(typeName, "Response") {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						for _, field := range structType.Fields.List {
							for _, ident := range field.Names {
								if sensitiveRegex.MatchString(ident.Name) {
									tagValue := ""
									if field.Tag != nil {
										tagValue = field.Tag.Value
									}
									if !strings.Contains(tagValue, `json:"-"`) {
										c.AddViolation(SevError, "Security.SensitiveFieldInDTO", relPath, field.Pos(),
											fmt.Sprintf("Sensitive field '%s' detected in public DTO struct '%s'!", ident.Name, typeName))
									}
								}
							}
						}
					}
				}
			}
			return true
		})

		return nil
	})
	c.Pass("Security.DTOCheck", "No sensitive credentials leaked in Response DTOs")
}

// 7. Check Error Handling & Panic Usage
func (c *Checker) checkErrorHandlingAndPanics() {
	fmt.Printf("%s[7/8] Checking Error Handling & Runtime Panic Rules...%s\n", colorBold, colorReset)

	_ = filepath.Walk(c.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || isGeneratedFile(path) {
			return nil
		}

		// Skip test files and cmd tools
		if strings.HasSuffix(path, "_test.go") || strings.Contains(path, filepath.Join("cmd", "tools")) {
			return nil
		}

		relPath, _ := filepath.Rel(c.rootDir, path)

		// Focus on internal/modules and internal/common
		if !strings.HasPrefix(relPath, filepath.Join("internal", "modules")) &&
			!strings.HasPrefix(relPath, filepath.Join("internal", "common")) {
			return nil
		}

		node, parseErr := parser.ParseFile(c.fset, path, nil, 0)
		if parseErr != nil || node == nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
					c.AddViolation(SevError, "ErrorHandling.NoPanic", relPath, call.Pos(),
						"Use explicit error returns (*common.AppError) instead of 'panic()' in runtime code")
				}
			}
			return true
		})

		return nil
	})
	c.Pass("ErrorHandling.NoPanic", "No unhandled panic calls found in business modules")
}

// 8. Check Dependency Injection & register.go Contracts
func (c *Checker) checkRegisterAndDIContracts() {
	fmt.Printf("%s[8/8] Checking DI ProviderSets & register.go Interface Assertions...%s\n", colorBold, colorReset)

	modulesDir := filepath.Join(c.rootDir, "internal", "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modName := entry.Name()
		registerFile := filepath.Join(modulesDir, modName, "register.go")
		if _, err := os.Stat(registerFile); os.IsNotExist(err) {
			continue
		}

		relPath := filepath.Join("internal", "modules", modName, "register.go")
		node, parseErr := parser.ParseFile(c.fset, registerFile, nil, 0)
		if parseErr != nil || node == nil {
			continue
		}

		hasProviderSet := false
		hasRegisterRoutesMethod := false
		hasAutoMigrateMethod := false
		hasRouteRegistrarAssertion := false
		hasMigratorAssertion := false

		ast.Inspect(node, func(n ast.Node) bool {
			// Check ProviderSet
			if valSpec, ok := n.(*ast.ValueSpec); ok {
				for _, name := range valSpec.Names {
					if name.Name == "ProviderSet" {
						hasProviderSet = true
					}
				}
			}

			// Check methods
			if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Recv != nil {
				if funcDecl.Name.Name == "RegisterRoutes" {
					hasRegisterRoutesMethod = true
				}
				if funcDecl.Name.Name == "AutoMigrate" {
					hasAutoMigrateMethod = true
				}
			}

			// Check interface assertions: var _ common.RouteRegistrar = (*Module)(nil)
			if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
				for _, spec := range genDecl.Specs {
					if valSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valSpec.Names {
							if name.Name == "_" {
								typeStr := fmt.Sprintf("%v", valSpec.Type)
								if strings.Contains(typeStr, "RouteRegistrar") {
									hasRouteRegistrarAssertion = true
								}
								if strings.Contains(typeStr, "Migrator") {
									hasMigratorAssertion = true
								}
							}
						}
					}
				}
			}

			return true
		})

		if !hasProviderSet {
			c.AddViolation(SevError, "DI.ProviderSetMissing", relPath, token.NoPos,
				fmt.Sprintf("Module '%s/register.go' must declare a Wire 'ProviderSet'", modName))
		}

		if hasRegisterRoutesMethod && !hasRouteRegistrarAssertion {
			c.AddViolation(SevWarning, "Contract.RouteRegistrarAssertion", relPath, token.NoPos,
				fmt.Sprintf("Module '%s' has RegisterRoutes() but lacks assertion: 'var _ common.RouteRegistrar = (*Module)(nil)'", modName))
		}

		if hasAutoMigrateMethod && !hasMigratorAssertion {
			c.AddViolation(SevWarning, "Contract.MigratorAssertion", relPath, token.NoPos,
				fmt.Sprintf("Module '%s' has AutoMigrate() but lacks assertion: 'var _ common.Migrator = (*Module)(nil)'", modName))
		}
	}

	c.Pass("DI.RegisterContracts", "Module registration contracts and Wire ProviderSets verified")
}

func (c *Checker) printSummary() {
	errorCount := 0
	warnCount := 0

	for _, v := range c.violations {
		if v.Severity == SevError {
			errorCount++
		} else {
			warnCount++
		}
	}

	fmt.Println()
	if len(c.violations) > 0 {
		fmt.Printf("%s%s-------------------- VIOLATIONS REPORT --------------------%s\n", colorBold, colorRed, colorReset)
		for _, v := range c.violations {
			loc := v.File
			if v.Line > 0 {
				loc = fmt.Sprintf("%s:%d", v.File, v.Line)
			}

			if v.Severity == SevError {
				fmt.Printf("%s[FAIL]%s %s%s%s [%s]\n       -> %s\n",
					colorRed, colorReset, colorBold, loc, colorReset, v.Rule, v.Message)
			} else {
				fmt.Printf("%s[WARN]%s %s%s%s [%s]\n       -> %s\n",
					colorYellow, colorReset, colorBold, loc, colorReset, v.Rule, v.Message)
			}
		}
		fmt.Println()
	}

	fmt.Printf("%s======================== SUMMARY ========================%s\n", colorBold, colorReset)
	fmt.Printf(" Passed Checks : %s%d%s\n", colorGreen, c.passCount, colorReset)
	fmt.Printf(" Errors        : %s%d%s\n", colorRed, errorCount, colorReset)
	fmt.Printf(" Warnings      : %s%d%s\n", colorYellow, warnCount, colorReset)
	fmt.Printf("%s=========================================================%s\n", colorBold, colorReset)

	if errorCount > 0 || (c.strict && warnCount > 0) {
		fmt.Printf("\n%s❌ Architecture & Coding Standard Check FAILED!%s\n\n", colorRed, colorReset)
		os.Exit(1)
	}

	fmt.Printf("\n%s✅ All Architecture & Coding Standard Checks PASSED!%s\n\n", colorGreen, colorReset)
	os.Exit(0)
}
