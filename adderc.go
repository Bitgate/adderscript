package main

import (
	"io/ioutil"
	"encoding/xml"
)

type xmlDoc struct {
	Functions []xmlFunction `xml:"function"`
}

type xmlFunction struct {
	Name       string             `xml:"name,attr"`
	Id         int                `xml:"id,attr"`
	ReturnType string             `xml:"return,attr"`
	Params     []xmlFunctionParam `xml:"params"`
}

type xmlFunctionParam struct {
	Type string `xml:"type,attr"`
}

func main() {
	data, err := ioutil.ReadFile("scripts/input.adr")
	if err != nil {
		panic(err)
	}

	text := string(data)
	tokens := ScanText(text)

	// Load runtime
	functionsFile, err := ioutil.ReadFile("scripts/runtime/runtime_functions.xml")
	if err != nil {
		panic("cannot load runtime functions file!")
	}

	decoded := new(xmlDoc)
	err = xml.Unmarshal(functionsFile, &decoded)

	if err != nil {
		panic("cannot decode xml function file: " + err.Error())
	}

	// Convert XML to array of native methods
	natives := make([]*NativeMethod, len(decoded.Functions))[:0]
	for _, v := range decoded.Functions {
		newNative := NativeMethod{
			name:   v.Name,
			opcode: v.Id,
		}

		natives = append(natives, &newNative)
	}

	ast := Parse(text, tokens)
	assembler := Assembler{nativeMethods: natives}
	assembler.AssembleProgram(ast)
}
