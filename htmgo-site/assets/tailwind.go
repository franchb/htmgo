package assets

import _ "embed"

//go:embed public/jetbrains-tailwind.json
var JetBrainsTailwindConfig string

//go:embed public/vscode-tailwind.json
var VSCodeTailwindConfig string
