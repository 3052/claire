package claire

import (
   "bytes"
   "fmt"
   "go/ast"
   "go/format"
   "go/parser"
   "go/scanner"
   "go/token"
   "html"
   "html/template"
   "strings"
)

var builtInTypes = map[string]struct{}{
   "bool": {}, "byte": {}, "complex64": {}, "complex128": {},
   "error": {}, "float32": {}, "float64": {}, "int": {}, "int8": {},
   "int16": {}, "int32": {}, "int64": {}, "rune": {}, "string": {},
   "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
   "uintptr": {},
}

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
