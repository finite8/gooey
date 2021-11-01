package core

import "reflect"

func GetRenderableForStructField(sf reflect.StructField) Renderable {
	return NewTextPrimitve(sf.Name)
}
