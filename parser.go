package claire

import (
   "bytes"
   "fmt"
   "go/ast"
   "go/doc"
   "go/format"
   "go/parser"
   "go/scanner"
   "go/token"
   "html"
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

// --- Internal Helpers ---

func formatAndHighlight(node ast.Node, fset *token.FileSet, typeNames map[string]struct{}) (template.HTML, error) {
   var buf bytes.Buffer
   if err := format.Node(&buf, fset, node); err != nil {
      return "", fmt.Errorf("failed to format node: %w", err)
   }

   sourceString := buf.String()
   const prefix = "package p\n\n"
   wrappedSource := prefix + sourceString

   fsetForHighlighting := token.NewFileSet()
   astFile, err := parser.ParseFile(fsetForHighlighting, "snippet.go", wrappedSource, 0)
   if err != nil || len(astFile.Decls) == 0 {
      return syntaxHighlight(sourceString, token.NewFileSet(), nil)
   }

   newRootNode := astFile.Decls[0]
   rawOffsets := collectTypeUsageOffsets(newRootNode, fsetForHighlighting, typeNames)
   adjustedOffsets := make(map[int]struct{})

   for offset := range rawOffsets {
      adjustedOffset := offset - len(prefix)
      if adjustedOffset >= 0 {
         adjustedOffsets[adjustedOffset] = struct{}{}
      }
   }

   return syntaxHighlight(sourceString, fsetForHighlighting, adjustedOffsets)
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

func collectTypeUsageOffsets(rootNode ast.Node, fset *token.FileSet, typeNames map[string]struct{}) map[int]struct{} {
   offsets := make(map[int]struct{})
   ast.Inspect(rootNode, func(n ast.Node) bool {
      if n == nil {
         return false
      }
      switch x := n.(type) {
      case *ast.ValueSpec:
         collectFromExpr(x.Type, fset, typeNames, offsets)
      case *ast.TypeSpec:
         collectFromExpr(x.Type, fset, typeNames, offsets)
      case *ast.Field:
         collectFromExpr(x.Type, fset, typeNames, offsets)
      }
      return true
   })
   return offsets
}

func collectFromFields(fields *ast.FieldList, fset *token.FileSet, typeNames map[string]struct{}, offsets map[int]struct{}) {
   if fields == nil {
      return
   }
   for _, field := range fields.List {
      collectFromExpr(field.Type, fset, typeNames, offsets)
   }
}

func collectFromExpr(expr ast.Expr, fset *token.FileSet, typeNames map[string]struct{}, offsets map[int]struct{}) {
   if expr == nil {
      return
   }
   switch x := expr.(type) {
   case *ast.Ident:
      if _, isType := typeNames[x.Name]; isType {
         file := fset.File(x.Pos())
         if file != nil {
            offsets[file.Offset(x.Pos())] = struct{}{}
         }
      }
   case *ast.StarExpr:
      collectFromExpr(x.X, fset, typeNames, offsets)
   case *ast.ArrayType:
      collectFromExpr(x.Elt, fset, typeNames, offsets)
   case *ast.MapType:
      collectFromExpr(x.Key, fset, typeNames, offsets)
      collectFromExpr(x.Value, fset, typeNames, offsets)
   case *ast.ChanType:
      collectFromExpr(x.Value, fset, typeNames, offsets)
   case *ast.FuncType:
      collectFromFields(x.Params, fset, typeNames, offsets)
      collectFromFields(x.Results, fset, typeNames, offsets)
   case *ast.StructType:
      collectFromFields(x.Fields, fset, typeNames, offsets)
   case *ast.InterfaceType:
      collectFromFields(x.Methods, fset, typeNames, offsets)
   case *ast.SelectorExpr:
      collectFromExpr(x.X, fset, typeNames, offsets)
   case *ast.IndexExpr:
      collectFromExpr(x.X, fset, typeNames, offsets)
      collectFromExpr(x.Index, fset, typeNames, offsets)
   case *ast.IndexListExpr:
      collectFromExpr(x.X, fset, typeNames, offsets)
      for _, index := range x.Indices {
         collectFromExpr(index, fset, typeNames, offsets)
      }
   }
}

// --- Syntax Highlighting Logic ---

var builtInTypes = map[string]struct{}{
   "bool": {}, "byte": {}, "complex64": {}, "complex128": {},
   "error": {}, "float32": {}, "float64": {}, "int": {}, "int8": {},
   "int16": {}, "int32": {}, "int64": {}, "rune": {}, "string": {},
   "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
   "uintptr": {},
}

func syntaxHighlight(source string, fset *token.FileSet, typeOffsets map[int]struct{}) (template.HTML, error) {
   if fset == nil {
      fset = token.NewFileSet()
   }
   file := fset.AddFile("", fset.Base(), len(source))
   var s scanner.Scanner
   s.Init(file, []byte(source), nil, scanner.ScanComments)

   var buf strings.Builder
   lastOffset := 0

   for {
      pos, tok, lit := s.Scan()
      if tok == token.EOF {
         break
      }

      offset := file.Offset(pos)
      if lastOffset < offset {
         buf.WriteString(html.EscapeString(source[lastOffset:offset]))
      }

      tokenText := lit
      if tokenText == "" {
         tokenText = tok.String()
      }

      escapedToken := html.EscapeString(tokenText)
      var tokenHTML string

      if tok == token.IDENT {
         _, isTypeOffset := typeOffsets[offset]
         _, isBuiltIn := builtInTypes[lit]

         if isTypeOffset {
            tokenHTML = fmt.Sprintf(`<a href="#%s">%s</a>`, escapedToken, escapedToken)
         } else if isBuiltIn {
            tokenHTML = fmt.Sprintf(`<span class="builtin">%s</span>`, escapedToken)
         } else {
            tokenHTML = escapedToken
         }
      } else if tok.IsKeyword() {
         tokenHTML = fmt.Sprintf(`<span class="keyword">%s</span>`, escapedToken)
      } else if tok == token.COMMENT {
         tokenHTML = fmt.Sprintf(`<span class="comment">%s</span>`, escapedToken)
      } else if tok == token.STRING {
         tokenHTML = fmt.Sprintf(`<span class="string">%s</span>`, escapedToken)
      } else {
         tokenHTML = escapedToken
      }

      buf.WriteString(tokenHTML)
      lastOffset = offset + len(tokenText)
   }

   if lastOffset < len(source) {
      buf.WriteString(html.EscapeString(source[lastOffset:]))
   }

   return template.HTML(buf.String()), nil
}
