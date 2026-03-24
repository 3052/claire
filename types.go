// types.go
package claire

import "html/template"

// FuncDoc holds documentation for a single function or method.
type FuncDoc struct {
   Name      string
   Doc       string
   Signature template.HTML
}

// TypeDoc holds documentation for a type definition and its methods.
type TypeDoc struct {
   Name       string
   Doc        string
   Definition template.HTML
   Constants  []VarDoc  // Constants grouped under this type
   Variables  []VarDoc  // Variables grouped under this type
   Functions  []FuncDoc // Constructors/factory functions associated with this type.
   Methods    []FuncDoc
}

// VarDoc holds documentation for a variable or constant declaration.
type VarDoc struct {
   Doc        string
   Definition template.HTML
}
