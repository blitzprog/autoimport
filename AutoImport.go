package autoimport

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/akyoto/autoimport/parser"
	"github.com/akyoto/color"
)

// AutoImport helps you automatically add imports to your source files.
type AutoImport struct {
	index           PackageIndex
	moduleDirectory string
}

// New creates a new auto import.
func New(moduleDirectory string) *AutoImport {
	standardPackagesPath := getStandardPackagesPath()
	standardPackages := getPackagesInDirectory(standardPackagesPath, standardPackagesPath)
	goModPath := findGoMod(moduleDirectory)
	dependencies, moduleImportPath, err := readGoMod(goModPath)

	if err != nil {
		panic(err)
	}

	// Find where modules are cached
	goModulesPath := getGoModulesPath()

	for _, dep := range dependencies {
		directoryName := fmt.Sprintf("%s@%s", dep.ImportPath, dep.Version)
		packageLocation := path.Join(goModulesPath, directoryName)
		importedPackages := getPackagesInDirectory(packageLocation, goModulesPath)
		merge(standardPackages, importedPackages)
	}

	// Local packages
	innerPackages := getPackagesInDirectory(moduleDirectory, moduleDirectory)

	for _, packageList := range innerPackages {
		for i := range packageList {
			packageList[i].ImportPath = fmt.Sprintf("%s/%s", moduleImportPath, packageList[i].ImportPath)
		}
	}

	merge(standardPackages, innerPackages)

	return &AutoImport{
		index:           standardPackages,
		moduleDirectory: moduleDirectory,
	}
}

// Imports returns the import paths used in the given source file.
func (importer *AutoImport) Imports(src []byte) []string {
	var imports []string
	identifiers := parser.PackageIdentifiers(src)

	for id := range identifiers {
		possiblePackages := importer.index[id]

		if len(possiblePackages) == 0 {
			continue
		}

		pkg := findCorrectPackage(possiblePackages)
		imports = append(imports, pkg.ImportPath)
	}

	// Sort by file system depth
	sort.Slice(imports, func(a int, b int) bool {
		countA := strings.Count(imports[a], "/")
		countB := strings.Count(imports[b], "/")

		if countA == countB {
			return imports[a] < imports[b]
		}

		return countA < countB
	})

	return imports
}

// Source finds the correct import statements and returns code that includes import paths.
func (importer *AutoImport) Source(src []byte) []byte {
	imports := importer.Imports(src)

	for _, importPath := range imports {
		fmt.Printf("%s\n", color.GreenString(importPath))
	}

	return src
}
