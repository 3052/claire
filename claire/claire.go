package main

import (
   "41.neocities.org/claire"
   "encoding/json"
   "errors"
   "flag"
   "log"
   "os"
   "path/filepath"
)

type flag_set struct {
   ImportPath string `json:"import_path"`
   Input      string
   Output     string
   Repository string
   Version    string
}

func (f *flag_set) do() error {
   log.Println("RemoveAll", f.Output)
   err := os.RemoveAll(f.Output)
   if err != nil {
      return err
   }
   err = claire.Generate(
      f.Input, f.Output, f.Repository, f.Version, f.ImportPath,
   )
   if err != nil {
      return err
   }
   // Check that both the HTML and CSS files were created.
   files_to_check := []string{
      "index.html",
      "style.css",
   }
   for _, file := range files_to_check {
      output_file := filepath.Join(f.Output, file)
      _, err = os.Stat(output_file)
      if err != nil {
         return err
      }
   }
   return nil
}

func (f *flag_set) New(name string) error {
   data, err := os.ReadFile(name)
   if err != nil {
      return err
   }
   err = json.Unmarshal(data, f)
   if err != nil {
      return err
   }
   if f.ImportPath == "" {
      return errors.New("import_path")
   }
   if f.Input == "" {
      return errors.New("input")
   }
   if f.Output == "" {
      return errors.New("output")
   }
   if f.Repository == "" {
      return errors.New("repository")
   }
   if f.Version == "" {
      return errors.New("version")
   }
   return nil
}

func main() {
   log.SetFlags(log.Ltime)
   name := flag.String("n", "", "name")
   flag.Parse()
   if *name != "" {
      var set flag_set
      err := set.New(*name)
      if err != nil {
         log.Fatal(err)
      }
      err = set.do()
      if err != nil {
         log.Fatal(err)
      }
   } else {
      flag.Usage()
   }
}
