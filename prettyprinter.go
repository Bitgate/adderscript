package main

import (
	"fmt"
	"text/tabwriter"
	"os"
	"strconv"
	"strings"
)

func (a *Assembler) PrettyPrint() {
	fmt.Println("")
	fmt.Println("Pretty print output:")
	fmt.Println("---------------------")
	fmt.Println("")

	fmt.Println("Defined methods:")
	for i, v := range a.methods {
		fmt.Printf("\t%d: %s (%d instructions)\n", i, v.name, len(v.instructions))

		for ii, vv := range v.variables {
			fmt.Printf("\t\tArgument %d: %s\n", ii, vv.name)
		}

		fmt.Println()
	}

	fmt.Println("Method code:")
	tw := tabwriter.NewWriter(os.Stdout, 8, 4, 2, '\t', 0)
	for i, v := range a.methods {
		fmt.Printf("\t%s (id %d with %d instructions)\n", v.name, i, len(v.instructions))

		for _, instr := range v.instructions {
			printInstruction(tw, a, v, instr)
		}

		tw.Flush()
		fmt.Println()
	}
}

func printInstruction(tw *tabwriter.Writer, a *Assembler, m *Method, ins *Instruction) {
	output := ""
	op := ins.opcode

	if op == op_pushconst {
		val := a.cpool.values[ins.i]
		desc := "unknown cpool value"

		if val.Type == VarTypeString {
			desc = "string " + strconv.Quote(val.Value.(string))
		} else if val.Type == VarTypeInt {
			desc = "int " + strconv.Itoa(val.Value.(int))
		} else if val.Type == VarTypeLong {
			desc = "long " + strconv.FormatInt(val.Value.(int64), 10)
		}

		output = fmt.Sprintf("PUSHCONST %d\t; %s", ins.i, desc)
	} else if op == op_nativecall {
		output = fmt.Sprintf("NATIVECALL %s\t; id %d", a.runtime.FindFunctionById(ins.i).Name, ins.i)
	} else if op == op_jmp {
		output = fmt.Sprintf("JMP %d\t", ins.i)
	} else if op == op_jz {
		output = fmt.Sprintf("JZ %d\t", ins.i)
	} else if op == op_getivar {
		output = fmt.Sprintf("GETIVAR %d\t", ins.i)

		// Document if they're parameters
		if ins.i < len(m.arguments) {
			output += "; parameter " + m.arguments[ins.i].name
		}
	} else if op == op_call {
		output = fmt.Sprintf("CALL %d\t", ins.i)
	} else if op == op_return {
		output = fmt.Sprintf("RETURN\t")
	}

	// Labels are a corner-case: we need to print that with a custom format
	if op == op_label {
		tw.Write([]byte(fmt.Sprintf("\tLABEL_%d:\t\n", ins.address)))
		return
	}

	if output == "" {
		tw.Write([]byte(fmt.Sprintf("\t%04d: OP_%d [%d %d %d]\t\n", ins.address, ins.opcode, ins.i, ins.l, ins.s)))
	} else {
		tw.Write([]byte(fmt.Sprintf("\t%04d: %s\n", ins.address, output)))
	}
}

func (t VariableType) String() string {
	if t == VarTypeString {
		return "string"
	} else if t == VarTypeInt {
		return "int"
	} else if t == VarTypeLong {
		return "long"
	} else if t == VarTypeBool {
		return "bool"
	} else if t == VarTypeVoid {
		return "void"
	} else if t == VarTypeUnresolved {
		return "unresolved"
	} else { // We have an else case for those that are unhandled in this function, but do exist.
		return "undefined"
	}
}

func TypeListToString(sep string, types ...VariableType) string {
	// Convert list to string representation first
	asStrings := make([]string, len(types))
	for i, v := range types {
		asStrings[i] = v.String()
	}

	return strings.Join(asStrings, sep)
}