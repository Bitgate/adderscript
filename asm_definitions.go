package main

type opcode int

const (
	op_ipush opcode = 0
	op_lpush = 1
	op_strpush = 2
	op_bpush = 3
	op_add = 4
	op_sub = 5
	op_mul = 6
	op_div = 7
	op_mod = 8
	op_jmp = 9
	op_getivar = 10
	op_setivar = 11
	op_return = 12
	op_cat = 13
	op_jz = 14
	op_eq = 15
	op_call = 16
	op_nativecall = 17
	op_farcall = 18
	op_const = 19
	op_returnvalue = 20
	op_getlvar = 21
	op_setlvar = 22
	op_getsvar = 23
	op_setsvar = 24
	op_getgi = 25
	op_setgi = 26
	op_bitor = 27
	op_bitand = 28
	op_neq = 29
	op_less = 30
	op_lesseq = 31
	op_more = 32
	op_moreeq = 33
	op_both = 34
	op_neither = 35
	op_nativecall_varargs = 36
	op_label = 255
)