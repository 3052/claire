package claire

import (
   "log"
   "os"
   "path/filepath"
)

// Generate creates HTML documentation and a corresponding CSS file for the Go
// package in sourceDir and writes them to outputDir. It includes metadata for
// the repository, version, and go-import path. The template and stylesheet
// are embedded in the binary.
func Generate(sourceDir, outputDir, repoURL, version, importPath, vcs string) error {
   pkgDoc, err := Parse(sourceDir, repoURL, version, importPath, vcs)
   if err != nil {
      return err
   }

   // Ensure the output directory exists.
   if err := os.MkdirAll(outputDir, 0755); err != nil {
      return err
   }

   // Render the HTML file using the embedded template.
   htmlOutputPath := filepath.Join(outputDir, "index.html")
   if err := Render(pkgDoc, htmlOutputPath); err != nil {
      return err
   }

   // Write the embedded CSS file to the output directory.
   cssDestPath := filepath.Join(outputDir, "style.css")
   log.Printf("Creating file: %s", cssDestPath)
   return os.WriteFile(cssDestPath, []byte(styleFile), 0644)
}
