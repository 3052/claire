package claire

import (
   "html/template"
   "log"
   "os"
   "path/filepath"
)

// Render generates the HTML documentation file using the embedded template.
func Render(pkgDoc *PackageDoc, outputPath string) error {
   // Parse the template directly from the embedded string variable.
   tmpl, err := template.New("doc").Parse(templateFile)
   if err != nil {
      return err
   }

   if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
      return err
   }

   log.Printf("Creating file: %s", outputPath)
   file, err := os.Create(outputPath)
   if err != nil {
      return err
   }
   defer file.Close()

   return tmpl.Execute(file, pkgDoc)
}
