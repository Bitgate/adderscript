package main

import "io/ioutil"

func main() {
	data, err := ioutil.ReadFile("scripts/input.adr")
	if err != nil {
		panic(err)
	}

	text := string(data)
	tokens := ScanText(text)

	natives := []*NativeMethod{
		{
			name:   "show_captioned_options",
			opcode: 1,
		},
	}

	ast := Parse(text, tokens)
	assembler := Assembler {nativeMethods: natives}
	assembler.AssembleProgram(ast)
}
