package main

import (
	"io/ioutil"
	"fmt"
	"os"
	"strings"
)

func main() {
	directory := os.Args[1]

	dataRt, err := ioutil.ReadFile(directory + "/runtime.arl")
	if err != nil {
		panic(fmt.Errorf("error loading runtime: %s", err))
	}

	runtime, err := ParseRuntime(string(dataRt))
	if err != nil {
		panic(fmt.Errorf("error parsing runtime: %s", err))
	}

	fmt.Printf("Loaded runtime with %d functions and %d listeners.\n", len(runtime.Functions), len(runtime.Listeners))
	compileRecursive(runtime, directory, "")
}

func compileRecursive(runtime *AdderRuntime, base string, dir string) {
	fmt.Printf("Compiling recursive: %s %s\n", base, dir)
	srcbase := base + "/src/" + dir
	entries, e := ioutil.ReadDir(srcbase)

	if e == nil {
		for _, v := range entries {
			if v.IsDir() {
				compileRecursive(runtime, base, dir + "/" + v.Name())
			} else {
				data, err := ioutil.ReadFile(srcbase + "/" + v.Name())
				if err != nil {
					panic(err)
				}

				text := string(data)
				tokens := ScanText(text)

				ast := Parse(text, tokens)
				assembler := Assembler{runtime: runtime}
				assembler.AssembleProgram(ast)
				assembler.PrettyPrint()

				os.MkdirAll(base + "/bin/" + dir, os.ModePerm)
				err = assembler.EncodeToFile(base + "/bin/" + dir + "/" + strings.Replace(v.Name(), ".adr", ".abf", -1))

				if err != nil {
					panic(err)
				}
			}
		}
	}
}
