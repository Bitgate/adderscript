package main

import (
	"io/ioutil"
	"encoding/xml"
)

type functionDoc struct {
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

type exportsDoc struct {
	Exports []xmlExport `xml:"export"`
}

type xmlExport struct {
	Name       string             `xml:"name,attr"`
	Id         int                `xml:"id,attr"`
}

func main() {
	data, err := ioutil.ReadFile("catatest/src/resource_dungeons.adr")
	if err != nil {
		panic(err)
	}

	text := string(data)
	tokens := ScanText(text)

	// Load runtime
	functionsFile, err := ioutil.ReadFile("catatest/runtime/functions.xml")
	if err != nil {
		panic("cannot load runtime functions file!")
	}

	decoded := new(functionDoc)
	err = xml.Unmarshal(functionsFile, &decoded)

	if err != nil {
		panic("cannot decode xml function file: " + err.Error())
	}

	exportsFile, err := ioutil.ReadFile("catatest/runtime/exports.xml")
	if err != nil {
		panic("cannot load runtime exports file!")
	}

	exports := new(functionDoc)
	err = xml.Unmarshal(exportsFile, &exports)

	if err != nil {
		panic("cannot decode xml exports file: " + err.Error())
	}

	// Convert XML to array of native methods
	natives := map[string]*NativeMethod {}
	for _, v := range decoded.Functions {
		newNative := NativeMethod{
			name:   v.Name,
			opcode: v.Id,
		}

		natives[v.Name] = &newNative
	}

	ast := Parse(text, tokens)
	assembler := Assembler{nativeMethods: natives}
	assembler.AssembleProgram(ast)
	assembler.PrettyPrint()
}
