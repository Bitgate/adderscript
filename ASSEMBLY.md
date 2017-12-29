# Assembly format
Adderscript compiles to a custom binary format. This binary format has a small set of instructions, simplified as much
as possible to make implementing Adderscript as easy as possible.

## Brief binary specification
A rough example of a C-like struct defining the different entries in the binary file:
```c
typedef struct adder_binary {
    uint8 bytecode_version;
    uint16 trigger_count;
    adder_trigger triggers[trigger_count];
    uint16 method_count;
    adder_method methods[method_count];
    
    int32 instr_count;
    adder_instr instructions[instr_count];
};

typedef struct adder_trigger {
    uint32 uid;
    uint32 address;
    uint8 value_count;
    adder_value values[value_count];
};

typedef struct adder_method {
    uint16 index;
    uint32 entry_address;
};

typedef struct adder_cpool {
    uint16 value_count;
    adder_value values[value_count];
};

typedef struct adder_instr {
    uint8 opcode;
    void *operand;
};

typedef struct adder_value {
    uint8 type;
    void *value;
};
```

## Instructions
| Mnemonic | Opcode | Operand | Description |
| -------- | ------ | ------- | ----------- |
| [PUSHCONST](#PUSHCONST) | 0x00   | int16    | Push constant from constant pool at [operand] |
| JMP | 0x01 | int32 | Jump to absolute address [operand] |
| GETLOCAL | 0x02 | int16 | Push local variable value to stack from index [operand] |
| SETLOCAL | 0x03 | int16 | Pop stack and store into local variable at index [operand] |
| RETURN | 0x04 | / | Exit stack frame or terminate script if last frame |
| JZ | 0x05 | int32 | Jump to absolute address [operand] if top of stack is 0 |
| EQ | 0x06 | / | Pop two int values, push value 1 if equal, value 0 if not |
| CALL | 0x07 | int32 | Call function at absolute address [operand], creating new frame |
| NATIVECALL | 0x08 | int16 | Calls a defined runtime function, manipulates stack as needed |

#### PUSHCONST
Pushes a constant from the constant pool at a given index to the stack. The value is taken from the constant pool 
at the index the operant value points to, and then pushed onto the stack.

