package fancy

import (
	"fmt"
	"html/template"
	"io"
	"reflect"

	"example.com/goboganui/pkg/register"
)

type TableComponent struct {
	dataGetter func(register.PageContext) (interface{}, error)
}

var tableTemplate = template.Must(template.New("Table").Parse(`
<table>
<th>
</th>
</table>`))

// simple table data structure
type TableData struct {
	Headers []interface{}
	Rows    [][]interface{}
}

func NewTableComponent(f func(register.PageContext) (interface{}, error)) Component {
	return &TableComponent{
		dataGetter: f,
	}
}

func ArrayToTable(arrayOfValues interface{}) *TableData {
	table := TableData{}
	rv := reflect.ValueOf(arrayOfValues)
	rt := reflect.TypeOf(arrayOfValues).Elem()
	// create all of the headers for our table
	for ix := 0; ix < rt.NumField(); ix++ {
		field := rt.Field(ix)
		hdr := GetRenderableForStructField(field)
		table.Headers = append(table.Headers, hdr)
	}

	// now lets do the rows
	for ix := 0; ix < rv.Len(); ix++ {
		currItem := rv.Index(ix)
		var currRow []interface{}
		for r_ix := 0; r_ix < rt.NumField(); r_ix++ {
			val := currItem.Field(r_ix)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}

			rndr_val := NewTextPrimitve(fmt.Sprintf("%v", val.Interface()))
			currRow = append(currRow, rndr_val)
		}
		table.Rows = append(table.Rows, currRow)
	}
	return &table

}

func (tc *TableComponent) OnRegister(ctx register.Registerer) {

}

func (tc *TableComponent) WriteContent(ctx register.PageContext, w io.Writer) {
	data, err := tc.dataGetter(ctx)
	if err != nil {
		// we need to handle this somehow
		WriteComponentError(ctx, tc, err, w)
		return
	}
	var table *TableData
	switch v := data.(type) {
	case TableData:
		table = &v
	case *TableData:
		table = v
	default:

		rv := reflect.ValueOf(data)
		rt := rv.Type()
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
			rv = rv.Elem()
		}
		// now we should have the right element
		switch rt.Kind() {
		case reflect.Array, reflect.Slice:
			table = ArrayToTable(rv.Interface())
		default:
			WriteComponentError(ctx, tc, fmt.Errorf("%t cannot be represented as a table", data), w)
			return
		}
	}
	// lets write the table parts
	io.WriteString(w, `<table><tr>`)
	WriteElements(ctx, "<th>", "</th>", w, table.Headers...)
	io.WriteString(w, `</tr>`)
	for _, row := range table.Rows {
		io.WriteString(w, `<tr>`)
		WriteElements(ctx, "<td>", "</td>", w, row...)
		io.WriteString(w, `</tr>`)
	}
	io.WriteString(w, `</table>`)

}
