// +build generate

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

var flags = []flag{
	{"bool", "Prod", "prod", false, "Production mode: hide errors, don't reload."},
	{"string", "Domain", "domain", "localhost:8081", "Base domain"},
	{"string", "DomainStatic", "domainstatic", "static.localhost:8081", "Domain to serve static files from."},
	{"string", "Listen", "listen", "localhost:8081", "Address to listen on."},
	{"string", "DBFile", "dbconnect", "db/goatcounter.sqlite3", "Database connection string."},
	{"string", "SMTP", "smtp", "smtps://b42bfac68fec83:f8dd7327e3e8b3@smtp.mailtrap.io:465", "SMTP connection string."},
	{"string", "Sentry", "sentry", "", "Sentry connection string"},
	{"string", "CPUProfile", "cpuprofile", "", "Write CPU profile to this file."},
	{"string", "MemProfile", "memprofile", "", "Write memory profile to this file."},
}

type flag struct {
	Type    string      // Go type name (e.g. bool, string, etc.)
	Name    string      // Variable name in the cfg package (e.g. Listen).
	Flag    string      // Flag name (e.g. "listen")
	Default interface{} // Default value.
	Help    string      // Help text.
}

func main() {
	longest := 0
	for _, f := range flags {
		if len(f.Name) > longest {
			longest = len(f.Name)
		}
	}

	var buf bytes.Buffer
	err := tpl.Execute(&buf, struct {
		Flags   []flag
		Longest int
	}{flags, longest})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "template: %s\n", err)
		os.Exit(1)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "gofmt: %s\n", err)
		os.Exit(1)
	}

	fmt.Print(string(src))
}

var tpl = template.Must(template.New("").
	Option("missingkey=error").
	Funcs(template.FuncMap{
		"ucfirst": func(s string) string { return strings.Title(s) },
		"pad":     func(s string, l int) string { return s + strings.Repeat(" ", l-len(s)) },
	}).
	Parse(`//go:generate sh -c "go run gen.go > cfg.go"

// Code generated by gen.go; DO NOT EDIT.

// Package cfg handles the application configuration.
package cfg

import (
    "flag"
    "fmt"
)

// Configuration variables.
var ({{range $f := .Flags}}
    {{$f.Name}} {{$f.Type}} // {{.Help}}{{end}}
)

// Set configuration variables from os.Args.
func Set() {{"{"}}{{range $f := .Flags}}
    flag.{{$f.Type|ucfirst}}Var(&{{$f.Name}}, "{{$f.Flag}}", {{printf "%#v" $f.Default}}, "{{$f.Help}}"){{end}}
    flag.Parse()
}

// Print out all configuration values.
func Print() {{"{"}}{{range $f := .Flags}}
    fmt.Printf("{{pad $f.Name $.Longest}}   %#v\n", {{$f.Name}}){{end}}
}
`))
