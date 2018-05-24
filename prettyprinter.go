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
	for i, v := range a.program.methods {
		fmt.Printf("\t%d: %s (%d instructions)\n", i, v.name, len(v.instructions))

		for ii, vv := range v.variables {
			fmt.Printf("\t\tArgument %d: %s\n", ii, vv.name)
		}

		fmt.Println()
	}

	fmt.Println("Method code:")
	tw := tabwriter.NewWriter(os.Stdout, 8, 4, 2, '\t', 0)
	for i, v := range a.program.methods {
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
	op := ins.Opcode

	if op == op_pushconst {
		val := a.cpool.values[ins.cpoolIndex]
		desc := "unknown cpool value"

		if val.Type == VarTypeString {
			desc = "string " + strconv.Quote(val.Value.(string))
		} else if val.Type == VarTypeInt {
			desc = "int " + strconv.Itoa(val.Value.(int))
		} else if val.Type == VarTypeLong {
			desc = "long " + strconv.FormatInt(val.Value.(int64), 10)
		}

		output = fmt.Sprintf("PUSHCONST %d\t; %s", ins.cpoolIndex, desc)
	} else if op == op_nativecall {
		fn := a.program.runtime.FindFunctionById(ins.cpoolIndex)
		args := make([]string, len(fn.Parameters))
		for i := range args {
			args[i] = fn.Parameters[i].Type.String() + " " + fn.Parameters[i].Name
		}

		output = fmt.Sprintf("NATIVECALL %s\t; id %d, %s(%s)", fn.Name, ins.cpoolIndex, fn.Name, strings.Join(args, ", "))
	} else if op == op_jmp {
		output = fmt.Sprintf("JMP %d\t", ins.cpoolIndex)
	} else if op == op_jz {
		output = fmt.Sprintf("JZ %d\t", ins.cpoolIndex)
	} else if op == op_getlocal {
		output = fmt.Sprintf("GETLOCAL %d\t", ins.cpoolIndex)

		// Document if they're parameters
		if ins.cpoolIndex < len(m.arguments) {
			output += "; parameter " + m.arguments[ins.cpoolIndex].name
		}
	} else if op == op_call {
		output = fmt.Sprintf("CALL %d\t", ins.cpoolIndex)
	} else if op == op_return {
		output = fmt.Sprintf("RETURN\t")
	} else if op == op_setlocal {
		output = fmt.Sprintf("SETLOCAL %d\t", ins.cpoolIndex)
	} else if op == op_eq {
		output = fmt.Sprintf("EQ\t")
	}

	// Labels are a corner-case: we need to print that with a custom format
	if op == op_label {
		tw.Write([]byte(fmt.Sprintf("\tLABEL_%d:\t\n", ins.address)))
		return
	}

	if output == "" {
		tw.Write([]byte(fmt.Sprintf("\t%04d: OP_%d [%d]\t\n", ins.address, ins.Opcode, ins.cpoolIndex)))
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