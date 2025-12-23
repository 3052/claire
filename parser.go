package claire

import (
   "fmt"
   "go/ast"
   "go/doc"
   "go/parser"
   "go/token"
   "html/template"
   "os"
   "path/filepath"
   "strings"
)

// Parse parses the Go package in the given directory and populates the PackageDoc.
// It does not populate metadata fields like RepositoryURL or Version.
func (pkgDoc *PackageDoc) Parse(inputPath string) error {
   fset := token.NewFileSet()
   files, err := parseGoFiles(fset, inputPath)
   if err != nil {
      return err
   }
   if len(files) == 0 {
      return fmt.Errorf("no Go source files found in directory: %s", inputPath)
   }
   p, err := doc.NewFromFiles(fset, files, "./")
   if err != nil {
      return fmt.Errorf("failed to create doc package: %w", err)
   }
   typeNames := make(map[string]struct{})
   for _, t := range p.Types {
      typeNames[t.Name] = struct{}{}
   }
   pkgDoc.Name = p.Name
   pkgDoc.Doc = p.Doc
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
   if pkgDoc.Functions, err = processFuncs(p.Funcs); err != nil {
      return err
   }
   for _, t := range p.Types {
      def, err := process(t.Decl)
      if err != nil {
         return err
      }
      typeDoc := TypeDoc{Name: t.Name, Doc: t.Doc, Definition: def}
      if typeDoc.Functions, err = processFuncs(t.Funcs); err != nil {
         return err
      }
      if typeDoc.Methods, err = processFuncs(t.Methods); err != nil {
         return err
      }
      pkgDoc.Types = append(pkgDoc.Types, typeDoc)
   }
   if pkgDoc.Constants, err = processValues(p.Consts); err != nil {
      return err
   }
   if pkgDoc.Variables, err = processValues(p.Vars); err != nil {
      return err
   }
   return nil
}

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
