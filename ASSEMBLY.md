# Assembly format
Adderscript compiles to a custom binary format. This binary format has a small set of instructions, simplified as much
as possible to make implementing Adderscript as easy as possible.

## Instructions
| Mnemonic | Opcode | Operand | Description |
| -------- | ------ | ------- | ----------- |
| PUSHCONST | 0x00   | int16    | Push constant from constant pool at [operand] |
| JMP | 0x01 | int32 | Jump to absolute address [operand] |
| GETLOCAL | 0x02 | int16 | Push local variable value to stack from index [operand] |
| SETLOCAL | 0x03 | int16 | Pop stack and store into local variable at index [operand] |
| RETURN | 0x04 | / | Exit stack frame or terminate script if last frame |
| JZ | 0x05 | int32 | Jump to absolute address [operand] if top of stack is 0 |
| EQ | 0x06 | / | Pop two int values, push value 1 if equal, value 0 if not |
| CALL | 0x07 | int32 | Call function at absolute address [operand], creating new frame |
| NATIVECALL | 0x08 | int16 | Calls a defined runtime function, manipulates stack as needed |