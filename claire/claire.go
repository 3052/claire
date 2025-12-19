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

func main() {
   log.SetFlags(log.Ltime)
   name := flag.String("n", "", "name")
   flag.Parse()
   if *name != "" {
      err := new(command).do(*name)
      if err != nil {
         log.Fatal(err)
      }
   } else {
      flag.Usage()
   }
}

type command struct {
   ImportPath string `json:"import_path"`
   Input      string
   Output     string
   Repository string
   Version    string
}

func (c *command) do(name string) error {
   data, err := os.ReadFile(name)
   if err != nil {
      return err
   }
   err = json.Unmarshal(data, c)
   if err != nil {
      return err
   }
   if c.ImportPath == "" {
      return errors.New("import_path")
   }
   if c.Input == "" {
      return errors.New("input")
   }
   if c.Output == "" {
      return errors.New("output")
   }
   if c.Repository == "" {
      return errors.New("repository")
   }
   if c.Version == "" {
      return errors.New("version")
   }
   log.Println("RemoveAll", c.Output)
   err = os.RemoveAll(c.Output)
   if err != nil {
      return err
   }
   err = claire.Generate(
      c.Input, c.Output, c.Repository, c.Version, c.ImportPath,
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
      output_file := filepath.Join(c.Output, file)
      _, err = os.Stat(output_file)
      if err != nil {
         return err
      }
   }
   return nil
}
