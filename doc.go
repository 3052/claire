package claire

import (
   "io/fs"
   "log"
   "os"
   "path" // Use the "path" package for URL-like paths
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

   // 2. Calculate the absolute path for the stylesheet link from the import path.
   styleSheetPath := calculateStyleSheetPath(importPath)

   // 3. Discover all package directories within the module.
   allPackagePaths, err := findAllPackageDirs(sourceDir)
   if err != nil {
      return err
   }

   var rootPackageExists bool
   var subPackagePaths []string
   for _, p := range allPackagePaths {
      if p == "." {
         rootPackageExists = true
      } else {
         subPackagePaths = append(subPackagePaths, p)
      }
   }
   sort.Strings(subPackagePaths)

   // 4. Generate docs for all sub-packages.
   var subPackageInfos []PackageInfo
   for _, pkgPath := range subPackagePaths {
      fullPath := filepath.Join(sourceDir, pkgPath)
      // CORRECTED: Use path.Join for URL-like import paths to ensure forward slashes.
      pkgImportPath := path.Join(importPath, filepath.ToSlash(pkgPath))
      pkgOutputDir := filepath.Join(outputDir, pkgPath)

      pkgDoc, err := Parse(fullPath, repoURL, version, pkgImportPath, styleSheetPath)
      if err != nil {
         log.Printf("Skipping directory %s: %v", fullPath, err)
         continue
      }

      if pkgDoc.IsEmpty() {
         log.Printf("Skipping empty package: %s", fullPath)
         continue
      }

      htmlOutputPath := filepath.Join(pkgOutputDir, "index.html")
      if err := Render(pkgDoc, htmlOutputPath); err != nil {
         return err
      }

      subPackageInfos = append(subPackageInfos, PackageInfo{
         Name: pkgDoc.Name,
         Path: filepath.ToSlash(pkgPath),
      })
   }

   // 5. Generate the root index.html.
   var rootDoc *PackageDoc
   if rootPackageExists {
      rootDoc, err = Parse(sourceDir, repoURL, version, importPath, styleSheetPath)
      if err != nil {
         return err
      }
   } else {
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
   // CORRECTED: Use path.Join for URL paths.
   return path.Join("/", pathPrefix, "style.css")
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
