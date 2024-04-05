package cli

var helpNameTemplate = `{{$v := offset .HelpName 6}}{{wrap .HelpName 3}}{{if .Usage}} - {{wrap .Usage $v}}{{end}}`
var usageTemplate = `{{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}} [command options] {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}`
var descriptionTemplate = `{{wrap .Description 3}}`
var visibleCommandTemplate = `{{ $cv := offsetCommands .VisibleCommands 5}}{{range .VisibleCommands}}
   {{$s := join .Names ", "}}{{$s}}{{ $sp := subtract $cv (offset $s 3) }}{{ indent $sp ""}}{{wrap .Usage $cv}}{{end}}`

var CommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{template "usageTemplate" .}}{{if .SubCommandsAvailable}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}
`

var AppHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{template "usageTemplate" .}}{{if .CommandVisible}}

COMMANDS:{{template "visibleCommandTemplate" .}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}
`
