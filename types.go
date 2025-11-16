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
   VCS            string
   StyleSheetPath string
   Doc            string
   Functions      []FuncDoc
   Types          []TypeDoc
   Variables      []VarDoc
   Constants      []VarDoc
   SubPackages    []PackageInfo
}
