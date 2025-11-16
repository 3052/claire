package claire

import (
   "io/fs"
   "log"
   "os"
   "path/filepath"
   "sort"
   "strings"
)

// Generate creates HTML documentation for all packages within a Go module.
func Generate(sourceDir, outputDir, repoURL, version, importPath string) error {
   // 1. Setup the output directory and the single stylesheet.
   if err := os.MkdirAll(outputDir, 0755); err != nil {
      return err
   }
   cssDestPath := filepath.Join(outputDir, "style.css")
   log.Printf("Creating file: %s", cssDestPath)
   if err := os.WriteFile(cssDestPath, []byte(styleFile), 0644); err != nil {
      return err
   }

   // 2. Prepare context for parsing runs.
   styleSheetPath := calculateStyleSheetPath(importPath)
   allPackagePaths, err := findAllPackageDirs(sourceDir)
   if err != nil {
      return err
   }

   // 3. Parse all discovered packages.
   parsedDocs := make(map[string]*PackageDoc)
   for _, pkgPath := range allPackagePaths {
      fullPath := filepath.Join(sourceDir, pkgPath)
      pkgImportPath := filepath.Join(importPath, pkgPath)
      pkgDoc, err := Parse(fullPath, repoURL, version, pkgImportPath, styleSheetPath)
      if err != nil {
         log.Printf("Skipping directory %s: %v", fullPath, err)
         continue
      }
      if !pkgDoc.IsEmpty() {
         parsedDocs[pkgPath] = pkgDoc
      } else {
         log.Printf("Skipping empty package: %s", fullPath)
      }
   }

   // 4. Prepare the list of sub-packages for the root index.
   var subPackageInfos []PackageInfo
   var subPackagePaths []string
   for pkgPath := range parsedDocs {
      if pkgPath != "." {
         subPackagePaths = append(subPackagePaths, pkgPath)
      }
   }
   sort.Strings(subPackagePaths)

   for _, pkgPath := range subPackagePaths {
      pkgDoc := parsedDocs[pkgPath]
      subPackageInfos = append(subPackageInfos, PackageInfo{
         Name: pkgDoc.Name,
         Path: filepath.ToSlash(pkgPath),
      })
   }

   // 5. Render all sub-package documentation files.
   for _, pkgPath := range subPackagePaths {
      pkgDoc := parsedDocs[pkgPath]
      htmlOutputPath := filepath.Join(outputDir, pkgPath, "index.html")
      if err := Render(pkgDoc, htmlOutputPath); err != nil {
         return err
      }
   }

   // 6. Prepare and render the root index.html.
   rootDoc, rootExists := parsedDocs["."]
   if !rootExists {
      rootDoc = &PackageDoc{
         Name:           filepath.Base(importPath),
         RepositoryURL:  repoURL,
         Version:        version,
         ImportPath:     importPath,
         StyleSheetPath: styleSheetPath,
      }
   }
   rootDoc.SubPackages = subPackageInfos
   indexPath := filepath.Join(outputDir, "index.html")
   return Render(rootDoc, indexPath)
}

// calculateStyleSheetPath determines the absolute URL path for the stylesheet.
func calculateStyleSheetPath(importPath string) string {
   var pathPrefix string
   if parts := strings.SplitN(importPath, "/", 2); len(parts) > 1 {
      pathPrefix = parts[1]
   }
   return filepath.ToSlash(filepath.Join("/", pathPrefix, "style.css"))
}

// findAllPackageDirs walks a directory and finds all subdirectories containing .go files.
func findAllPackageDirs(root string) ([]string, error) {
   packageSet := make(map[string]struct{})
   err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
      if err != nil {
         return err
      }
      if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
         dir := filepath.Dir(path)
         relDir, _ := filepath.Rel(root, dir)
         packageSet[relDir] = struct{}{}
      }
      return nil
   })
   packages := make([]string, 0, len(packageSet))
   for pkg := range packageSet {
      packages = append(packages, pkg)
   }
   return packages, err
}
