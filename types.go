package claire

import "html/template"

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
   Functions  []FuncDoc // Constructors/factory functions associated with this type.
   Methods    []FuncDoc
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
