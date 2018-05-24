package main

import (
	"io/ioutil"
	"bytes"
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const AbiVersion = 4

func (a *Assembler) Encode() []byte {
	buffer := new(bytes.Buffer)
	writer := bufio.NewWriter(buffer)

	writer.WriteByte(AbiVersion)

	// Encode triggers/event listeners
	// TODO make triggers listeners on strings too. And support wildcards.
	binary.Write(writer, binary.BigEndian, uint16(len(a.program.triggers)))
	for _, trigger := range a.program.triggers {
		binary.Write(writer, binary.BigEndian, int32(trigger.definition.InternalId))
		binary.Write(writer, binary.BigEndian, int32(trigger.label.address))

		// Encode the trigger value
		binary.Write(writer, binary.BigEndian, int8(len(trigger.values)))
		for _, v := range trigger.values {
			switch x := v.(type) {
			case int:
			case int32:
			case uint32:
				encodeAdderValue(writer, VarTypeInt, int32(x))
			case int64:
				encodeAdderValue(writer, VarTypeLong, x)
			case string:
				encodeAdderValue(writer, VarTypeString, x)
			default:
				panic(fmt.Errorf("cannot serialize type %T into a listener value", v))
			}
		}
	}

	// Encode methods..
	numInstructions := 0
	binary.Write(writer, binary.BigEndian, uint16(len(a.program.methods)))
	for _, method := range a.program.methods {
		binary.Write(writer, binary.BigEndian, int16(method.index))
		binary.Write(writer, binary.BigEndian, int32(method.entry.address))

		for _, inst := range method.instructions {
			if inst.Opcode != op_label {
				numInstructions++
			}
		}
	}

	// Encode constant pool
	binary.Write(writer, binary.BigEndian, int16(len(a.cpool.values)))
	for _, v := range a.cpool.values {
		encodeAdderValue(writer, v.Type, v.Value)
	}

	// Encode actual method code
	binary.Write(writer, binary.BigEndian, int32(numInstructions))
	for _, method := range a.program.methods {
		for _, inst := range method.instructions {
			if inst.Opcode != op_label {
				binary.Write(writer, binary.BigEndian, int8(inst.Opcode))

				if inst.Opcode == op_pushconst || inst.Opcode == op_nativecall ||
					inst.Opcode == op_setlocal || inst.Opcode == op_getlocal {
					binary.Write(writer, binary.BigEndian, int16(inst.cpoolIndex))
				} else if inst.Opcode == op_call || inst.Opcode == op_jz || inst.Opcode == op_jmp {
					binary.Write(writer, binary.BigEndian, int32(inst.cpoolIndex))
				}
			}
		}
	}

	writer.Flush()
	return buffer.Bytes()
}

func encodeAdderValue(w io.Writer, typ VariableType, value interface{}) {
	if typ == VarTypeInt {
		binary.Write(w, binary.BigEndian, int8(0))
		binary.Write(w, binary.BigEndian, int32(value.(int)))
	} else if typ == VarTypeLong {
		binary.Write(w, binary.BigEndian, int8(1))
		binary.Write(w, binary.BigEndian, value.(int64))
	} else if typ == VarTypeString {
		str := []byte(value.(string))

		binary.Write(w, binary.BigEndian, int8(2))
		binary.Write(w, binary.BigEndian, uint16(len(str)))
		binary.Write(w, binary.BigEndian, str)
	} else {
		panic("cannot encode type " + typ.String())
	}
}

func (a *Assembler) EncodeToFile(file string) error {
	data := a.Encode()
	return ioutil.WriteFile(file, data, 0664)
}