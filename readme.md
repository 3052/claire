# claire

> Every second counts
>
> [Clairo](//youtube.com/watch?v=CeA92xqw-QI) (2019)

documentation generator

---

Go language, I would like to create a package that creates HTML documentation
for a Go package

1. package will be called "claire"
2. do not include any "main.go" or "package main"
3. use a separate file for each type
4. any templates should be a separate file not a string
5. put all package files in the top directory not a subfolder
6. exclude unexported items
7. include a test file in the top folder
8. test file should only use the "example" directory, assuming the user has
   provided it, not create it
9. test output should remain after test is complete
10. do not include any other test files
11. when sending updates, send the complete file, for only new or updated
   files

## prior art

1. https://github.com/golang/go/issues/2381
2. https://github.com/abhinav/doc2go/issues/74
3. https://github.com/JohnStarich/go/issues/61
4. https://codeberg.org/pfad.fr/vanitydoc/issues/21
5. https://github.com/Vanilla-OS/Pallas/issues/10
6. https://github.com/viamrobotics/govanity/issues/6
7. https://github.com/dsnet/godoc/issues/3
8. https://github.com/goradd/moddoc/issues/2
