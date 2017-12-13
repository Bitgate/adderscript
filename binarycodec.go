package main

import (
	"io/ioutil"
	"bytes"
	"bufio"
	"encoding/binary"
	"strconv"
)

const AbiVersion = 3

func (a *Assembler) Encode() []byte {
	buffer := new(bytes.Buffer)
	writer := bufio.NewWriter(buffer)

	writer.WriteByte(AbiVersion)

	// Encode triggers/event listeners
	// TODO make triggers listeners on strings too. And support wildcards.
	binary.Write(writer, binary.BigEndian, uint16(len(a.triggers)))
	for _, trigger := range a.triggers {
		binary.Write(writer, binary.BigEndian, int32(trigger.definition.InternalId))
		binary.Write(writer, binary.BigEndian, int32(trigger.label.address))

		// Encode the trigger value
		binary.Write(writer, binary.BigEndian, int8(len(trigger.values)))
		binary.Write(writer, binary.BigEndian, int64(trigger.value))
	}

	// Encode methods..
	numInstructions := 0
	binary.Write(writer, binary.BigEndian, uint16(len(a.methods)))
	for _, method := range a.methods {
		binary.Write(writer, binary.BigEndian, int16(method.index))
		binary.Write(writer, binary.BigEndian, int32(method.entry.address))

		for _, inst := range method.instructions {
			if inst.opcode != op_label {
				numInstructions++
			}
		}
	}

	// Encode constant pool
	binary.Write(writer, binary.BigEndian, int16(len(a.cpool.values)))
	for _, v := range a.cpool.values {
		binary.Write(writer, binary.BigEndian, uint8(v.Type))

		if v.Type == VarTypeInt {
			binary.Write(writer, binary.BigEndian, int32(v.Value.(int)))
		} else if v.Type == VarTypeLong {
			binary.Write(writer, binary.BigEndian, v.Value.(int64))
		} else if v.Type == VarTypeString {
			str := []byte(v.Value.(string))

			binary.Write(writer, binary.BigEndian, uint16(len(str)))
			binary.Write(writer, binary.BigEndian, str)
		} else {
			panic("cannot encode type " + strconv.Itoa(int(v.Type)))
		}
	}

	// Encode actual method code
	binary.Write(writer, binary.BigEndian, int32(numInstructions))
	for _, method := range a.methods {
		for _, inst := range method.instructions {
			if inst.opcode != op_label {
				binary.Write(writer, binary.BigEndian, int8(inst.opcode))

				if inst.opcode == op_pushconst || inst.opcode == op_nativecall {
					binary.Write(writer, binary.BigEndian, int16(inst.i))
				} else if inst.opcode == op_call {
					binary.Write(writer, binary.BigEndian, int32(inst.i))
				}
			}
		}
	}

	writer.Flush()
	return buffer.Bytes()
}

func (a *Assembler) EncodeToFile(file string) error {
	data := a.Encode()
	return ioutil.WriteFile(file, data, 0664)
}