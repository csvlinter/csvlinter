//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	csvlinter "github.com/csvlinter/csvlinter/pkg/csvlinter"
)

func validateCSVJS(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return jsError("validateCSV requires 2 arguments: csvData (Uint8Array), options (object)")
	}

	csvData := make([]byte, args[0].Length())
	js.CopyBytesToGo(csvData, args[0])

	optsJS := args[1]
	opts := csvlinter.Options{
		Format: "json",
	}

	if v := optsJS.Get("filename"); !v.IsUndefined() && !v.IsNull() {
		opts.Filename = v.String()
	}
	if v := optsJS.Get("delimiter"); !v.IsUndefined() && !v.IsNull() && v.String() != "" {
		opts.Delimiter = v.String()
	}
	if v := optsJS.Get("failFast"); !v.IsUndefined() && !v.IsNull() {
		opts.FailFast = v.Bool()
	}
	if v := optsJS.Get("inferSchema"); !v.IsUndefined() && !v.IsNull() {
		opts.InferSchema = v.Bool()
	}
	if v := optsJS.Get("inferSchemaMaxRows"); !v.IsUndefined() && !v.IsNull() {
		opts.InferSchemaMaxRows = v.Int()
	}
	if v := optsJS.Get("schemaContent"); !v.IsUndefined() && !v.IsNull() && v.Length() > 0 {
		schemaData := make([]byte, v.Length())
		js.CopyBytesToGo(schemaData, v)
		opts.SchemaReader = bytes.NewReader(schemaData)
	}

	var out bytes.Buffer
	results, err := csvlinter.LintAdvanced(bytes.NewReader(csvData), opts, &out)
	if err != nil {
		return jsError(err.Error())
	}

	b, err := json.Marshal(results)
	if err != nil {
		return jsError(err.Error())
	}

	return js.Global().Get("JSON").Call("parse", string(b))
}

func jsError(msg string) js.Value {
	obj := js.Global().Get("Object").New()
	obj.Set("error", msg)
	return obj
}

func main() {
	js.Global().Set("csvlinterValidate", js.FuncOf(validateCSVJS))
	// Block forever — the Go runtime must stay alive to serve JS calls.
	select {}
}
