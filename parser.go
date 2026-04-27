// parser.go
package claire

import (
   "fmt"
   "go/ast"
   "go/doc"
   "go/parser"
   "go/token"
   "html/template"
   "log"
   "os"
   "path/filepath"
   "strings"
)

func parseGoFiles(fset *token.FileSet, dir string) ([]*ast.File, error) {
   entries, err := os.ReadDir(dir)
   if err != nil {
      return nil, err
   }
   var files []*ast.File
   var packageName string
   for _, entry := range entries {
      if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
         continue
      }
      path := filepath.Join(dir, entry.Name())
      file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
      if err != nil {
         return nil, err
      }
      if packageName == "" {
         packageName = file.Name.Name
      } else if file.Name.Name != packageName {
         return nil, fmt.Errorf("multiple package names found in directory: %s and %s", packageName, file.Name.Name)
      }
      files = append(files, file)
   }
   return files, nil
}

// PackageDoc holds all the documentation for a single package.
type PackageDoc struct {
   Name           string
   RepositoryUrl  string
   Version        string
   ImportPath     string
   StyleSheetPath string
   Doc            string
   Functions      []FuncDoc
   Types          []TypeDoc
   Variables      []VarDoc
   Constants      []VarDoc
   SubPackages    []string
}

// IsEmpty reports whether the package documentation is empty (has no content).
func (p *PackageDoc) IsEmpty() bool {
   return p.Doc == "" &&
      len(p.Constants) == 0 &&
      len(p.Variables) == 0 &&
      len(p.Functions) == 0 &&
      len(p.Types) == 0
}

// ParsePackageDoc parses the Go package in the given directory and returns an initialized PackageDoc.
// It does not populate metadata fields like RepositoryUrl or Version.
func ParsePackageDoc(inputPath string) (*PackageDoc, error) {
   fset := token.NewFileSet()
   files, err := parseGoFiles(fset, inputPath)
   if err != nil {
      return nil, err
   }
   if len(files) == 0 {
      return nil, fmt.Errorf("no Go source files found in directory: %s", inputPath)
   }
   docPkg, err := doc.NewFromFiles(fset, files, "./")
   if err != nil {
      return nil, fmt.Errorf("failed to create doc package: %w", err)
   }
   typeNames := make(map[string]struct{})
   for _, t := range docPkg.Types {
      typeNames[t.Name] = struct{}{}
   }

   pkgDoc := &PackageDoc{
      Name: docPkg.Name,
      Doc:  docPkg.Doc,
   }

   // -- Helpers to reduce boilerplate --
   process := func(decl ast.Decl) (template.HTML, error) {
      return formatAndHighlight(decl, fset, typeNames)
   }

   processFuncs := func(funcs []*doc.Func) ([]FuncDoc, error) {
      var docs []FuncDoc
      for _, f := range funcs {
         sig, err := process(f.Decl)
         if err != nil {
            return nil, err
         }
         docs = append(docs, FuncDoc{Name: f.Name, Doc: f.Doc, Signature: sig})
      }
      return docs, nil
   }

   processValues := func(values []*doc.Value) ([]VarDoc, error) {
      var docs []VarDoc
      for _, v := range values {
         def, err := process(v.Decl)
         if err != nil {
            return nil, err
         }
         docs = append(docs, VarDoc{Doc: v.Doc, Definition: def})
      }
      return docs, nil
   }

   // -- Processing --
   if pkgDoc.Functions, err = processFuncs(docPkg.Funcs); err != nil {
      return nil, err
   }

   for _, t := range docPkg.Types {
      def, err := process(t.Decl)
      if err != nil {
         return nil, err
      }
      typeDoc := TypeDoc{Name: t.Name, Doc: t.Doc, Definition: def}

      // Map constants and variables explicitly grouped under this type
      if typeDoc.Constants, err = processValues(t.Consts); err != nil {
         return nil, err
      }
      if typeDoc.Variables, err = processValues(t.Vars); err != nil {
         return nil, err
      }

      if typeDoc.Functions, err = processFuncs(t.Funcs); err != nil {
         return nil, err
      }
      if typeDoc.Methods, err = processFuncs(t.Methods); err != nil {
         return nil, err
      }
      pkgDoc.Types = append(pkgDoc.Types, typeDoc)
   }

   // Map untyped / base-level constants and variables
   if pkgDoc.Constants, err = processValues(docPkg.Consts); err != nil {
      return nil, err
   }
   if pkgDoc.Variables, err = processValues(docPkg.Vars); err != nil {
      return nil, err
   }

   return pkgDoc, nil
}

// Render generates the HTML documentation file using the embedded template.
func (p *PackageDoc) Render(outputPath string) error {
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
   return tmpl.Execute(file, p)
}
