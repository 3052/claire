// types.go
package claire

import (
   _ "embed"
   "html/template"
)

//go:embed package.tmpl
var packageTemplateFile string

//go:embed style.css
var styleFile string

// FuncDoc holds documentation for a single function or method.
type FuncDoc struct {
   Name      string
   Doc       string
   Signature template.HTML
}

// TypeDoc holds documentation for a type definition and its methods.
type TypeDoc struct {
   Constants  []VarDoc // Constants grouped under this type
   Definition template.HTML
   Doc        string
   Functions  []FuncDoc // Constructors/factory functions associated with this type.
   Methods    []FuncDoc
   Name       string
   Variables  []VarDoc // Variables grouped under this type
}

// VarDoc holds documentation for a variable or constant declaration.
type VarDoc struct {
   Doc        string
   Definition template.HTML
}
