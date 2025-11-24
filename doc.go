package claire

import (
   "io/fs"
   "log"
   "os"
   "path"
   "path/filepath"
   "sort"
   "strings"
)

// Generate creates HTML documentation for all packages within a Go module.
func Generate(sourceDir, outputDir, repoURL, version, importPath string) error {
   if err := os.MkdirAll(outputDir, 0755); err != nil {
      return err
   }
   cssDestPath := filepath.Join(outputDir, "style.css")
   log.Printf("Creating file: %s", cssDestPath)
   if err := os.WriteFile(cssDestPath, []byte(styleFile), 0644); err != nil {
      return err
   }

   styleSheetPath := calculateStyleSheetPath(importPath)
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

   var subPackageInfos []PackageInfo
   for _, pkgPath := range subPackagePaths {
      fullPath := filepath.Join(sourceDir, pkgPath)
      pkgOutputDir := filepath.Join(outputDir, pkgPath)

      // Parse the package without metadata first.
      pkgDoc, err := Parse(fullPath)
      if err != nil {
         log.Printf("Skipping directory %s: %v", fullPath, err)
         continue
      }

      if pkgDoc.IsEmpty() {
         log.Printf("Skipping empty package: %s", fullPath)
         continue
      }

      // Inject metadata.
      pkgDoc.RepositoryURL = repoURL
      pkgDoc.Version = version
      pkgDoc.StyleSheetPath = styleSheetPath
      pkgDoc.ImportPath = path.Join(importPath, filepath.ToSlash(pkgPath))

      htmlOutputPath := filepath.Join(pkgOutputDir, "index.html")

      // CORRECTED: Calling Render as a method
      if err := pkgDoc.Render(htmlOutputPath); err != nil {
         return err
      }

      subPackageInfos = append(subPackageInfos, PackageInfo{
         Name: pkgDoc.Name,
         Path: filepath.ToSlash(pkgPath),
      })
   }

   var rootDoc *PackageDoc
   if rootPackageExists {
      rootDoc, err = Parse(sourceDir)
      if err != nil {
         return err
      }
   } else {
      rootDoc = &PackageDoc{Name: filepath.Base(importPath)}
   }

   // Inject metadata for the root package.
   rootDoc.RepositoryURL = repoURL
   rootDoc.Version = version
   rootDoc.ImportPath = importPath
   rootDoc.StyleSheetPath = styleSheetPath
   rootDoc.SubPackages = subPackageInfos

   indexPath := filepath.Join(outputDir, "index.html")

   // CORRECTED: Calling Render as a method
   return rootDoc.Render(indexPath)
}

func calculateStyleSheetPath(importPath string) string {
   var pathPrefix string
   if parts := strings.SplitN(importPath, "/", 2); len(parts) > 1 {
      pathPrefix = parts[1]
   }
   return path.Join("/", pathPrefix, "style.css")
}

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
