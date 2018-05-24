package main


type ConstantPool struct {
	values []*ConstantPoolEntry
}

type ConstantPoolEntry struct {
	Type VariableType
	Value interface{}
}

func (c *ConstantPool) getInt(i int) int {
	for k, v := range c.values {
		if v.Type == VarTypeInt && v.Value.(int) == i {
			return k
		}
	}

	c.values = append(c.values, &ConstantPoolEntry{
		Type:  VarTypeInt,
		Value: i,
	})

	return len(c.values) - 1
}

func (c *ConstantPool) getLong(i int64) int {
	for k, v := range c.values {
		if v.Type == VarTypeLong && v.Value.(int64) == i {
			return k
		}
	}

	c.values = append(c.values, &ConstantPoolEntry{
		Type:  VarTypeLong,
		Value: i,
	})

	return len(c.values) - 1
}

func (c *ConstantPool) getString(s string) int {
	for k, v := range c.values {
		if v.Type == VarTypeString && v.Value.(string) == s {
			return k
		}
	}

	c.values = append(c.values, &ConstantPoolEntry{
		Type:  VarTypeString,
		Value: s,
	})

	return len(c.values) - 1
}
