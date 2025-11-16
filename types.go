package claire

import (
   "html/template"
   "log"
   "os"
   "path/filepath"
   _ "embed"
)

//go:embed package.tmpl
var packageTemplateFile string

//go:embed style.css
var styleFile string

// Render generates the HTML documentation file using the embedded template.
func Render(pkgDoc *PackageDoc, outputPath string) error {
   // Parse the template directly from the embedded string variable.
   tmpl, err := template.New("package").Parse(packageTemplateFile)
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

// FuncDoc holds documentation for a single function or method.
type FuncDoc struct {
   Name      string
   Doc       string
   Signature template.HTML
}

// VarDoc holds documentation for a variable or constant declaration.
type VarDoc struct {
   Doc        string
   Definition template.HTML
}

// TypeDoc holds documentation for a type definition and its methods.
type TypeDoc struct {
   Name       string
   Doc        string
   Definition template.HTML
   Methods    []FuncDoc
}

// PackageInfo holds metadata for a single sub-package within a module.
type PackageInfo struct {
   Name string
   Path string
}

// PackageDoc holds all the documentation for a single package.
type PackageDoc struct {
   Name           string
   RepositoryURL  string
   Version        string
   ImportPath     string
   StyleSheetPath string
   Doc            string
   Functions      []FuncDoc
   Types          []TypeDoc
   Variables      []VarDoc
   Constants      []VarDoc
   SubPackages    []PackageInfo
}

// IsEmpty reports whether the package documentation is empty (has no content).
func (p *PackageDoc) IsEmpty() bool {
   return p.Doc == "" &&
      len(p.Constants) == 0 &&
      len(p.Variables) == 0 &&
      len(p.Functions) == 0 &&
      len(p.Types) == 0
}
