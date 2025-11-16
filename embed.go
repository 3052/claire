package claire

import _ "embed"

//go:embed template.tmpl
var templateFile string

//go:embed style.css
var styleFile string
