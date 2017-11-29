package main

import (
	"io/ioutil"
	"bytes"
	"bufio"
	"encoding/binary"
)

const AbiVersion = 3

func (a *Assembler) Encode() []byte {
	buffer := new(bytes.Buffer)
	writer := bufio.NewWriter(buffer)

	writer.WriteByte(AbiVersion)

	// Encode triggers/event listeners
	binary.Write(writer, binary.BigEndian, uint16(0))

	writer.Flush()
	return buffer.Bytes()
}

func (a *Assembler) EncodeToFile(file string) error {
	data := a.Encode()
	return ioutil.WriteFile(file, data, 0664)
}