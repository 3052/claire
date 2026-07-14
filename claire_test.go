package claire

import (
   "go/parser"
   "go/token"
   "testing"
)

func TestVariadicTypeNotCollected(t *testing.T) {
   src := "package p\n\nfunc Foo(args ...Bar) {}\n"
   fset := token.NewFileSet()
   file, err := parser.ParseFile(fset, "test.go", src, 0)
   if err != nil {
      t.Fatal(err)
   }

   typeNames := map[string]struct{}{"Bar": {}}
   offsets := collectTypeUsageOffsets(file, fset, typeNames)

   if len(offsets) == 0 {
      t.Fatal("expected Bar to be collected from variadic parameter, got 0 offsets")
   }
}
