package main

var TypeInt = BaseType{Name: "int", BuiltIn: true}
var TypeString = BaseType{Name: "string", BuiltIn: true}
var TypeVoid = BaseType{Name: "void", BuiltIn: true}

var TypePlayer = BaseType{
	Name: "Player",
	Fields: []TypeField{
		{
			Name: "name",
			Type: TypeString,
		},
	},
	Methods: []TypeMethod{
		// void teleport(int x, int y)
		{
			Name:    "teleport",
			Returns: TypeVoid,
			Parameters: []MethodParameter{
				{
					Name: "x",
					Type: TypeInt,
				},
				{
					Name: "z",
					Type: TypeInt,
				},
			},
		},
	},
}

type BaseType struct {
	Name    string
	Fields  []TypeField
	Methods []TypeMethod

	// BuiltIn indicates whether this type is defined by the language instead of by the host (or the runtime).
	// Types that are BuiltIn are mostly the basic primitives, the string type and perhaps a few other basic types.
	// If a type is BuiltIn and has no Fields or Methods, it is classified as a primitive.
	BuiltIn bool

	// Native indicates whether this type is handled by the host engine. If the type is native, all get and set
	// operations are performed through mapped get/set operations instead of direct field mutations. This also means
	// that a native type can have fields mapped to methods - eg. field "name" being used as a field where it really
	// is mapped to a getter/setter pair. This also allows for prettier 'computed values' - eg. player.ready and the likes.
	Native bool
}

// IsPrimitive returns true if the type is seen as primitive (no methods and no fields) and false if not.
// This only applies to BuiltIn types, not to user-defined or native types.
func (t BaseType) IsPrimitive() bool {
	return t.BuiltIn && t.Fields == nil && t.Methods == nil
}

func (t BaseType) ResolveField(name string) *TypeField {
	// Don't lookup when there are no fields.
	if t.Fields == nil {
		return nil
	}

	for _, v := range t.Fields {
		if v.Name == name {
			return &v
		}
	}

	return nil
}

type TypeField struct {
	Type BaseType
	Name string
}

type TypeMethod struct {
	Name       string
	Parameters []MethodParameter
	Returns    BaseType
}

type MethodParameter struct {
	Name string
	Type BaseType
}
