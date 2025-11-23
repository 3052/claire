package claire

import _ "embed"

//go:embed package.tmpl
var packageTemplateFile string

//go:embed style.css
var styleFile string
