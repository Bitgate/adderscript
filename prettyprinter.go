package main

import (
	"fmt"
	"text/tabwriter"
	"os"
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

	if op == op_ipush {
		output = fmt.Sprintf("IPUSH %d\t", a.cpool.ints[ins.i])
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