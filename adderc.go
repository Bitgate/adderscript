package main

import (
	"io/ioutil"
	"fmt"
)

func main() {
	dataRt, err := ioutil.ReadFile("scripts/runtime.arl")
	runtime, err := ParseRuntime(string(dataRt))
	if err != nil {
		panic(fmt.Errorf("error parsing runtime: %s", err))
	}

	fmt.Printf("Loaded runtime with %d functions and %d listeners.\n", len(runtime.Functions), len(runtime.Listeners))

	data, err := ioutil.ReadFile("scripts/input.adr")
	if err != nil {
		panic(err)
	}

	text := string(data)
	tokens := ScanText(text)

	ast := Parse(text, tokens)
	assembler := Assembler{runtime: runtime}
	assembler.AssembleProgram(ast)
	assembler.PrettyPrint()
}
