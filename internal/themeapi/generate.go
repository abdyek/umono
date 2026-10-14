// Package themeapi keeps the Theme API range of umono.json in step with
// themecore. The runtime reads the range from themecore itself; umono.json is
// for readers that do not run it, such as the CLI's install and upgrade.
package themeapi

//go:generate go run ./gen ../../umono.json
//go:generate go run ./domgen testdata/dom
