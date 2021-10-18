package fancy

import "reflect"

func GetRenderableForStructField(sf reflect.StructField) Renderable {
	return NewTextPrimitve(sf.Name)
}
