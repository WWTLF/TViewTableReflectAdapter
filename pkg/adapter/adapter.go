package adapter

import (
	"fmt"
	"reflect"

	"github.com/rivo/tview"
)

type Resolver func(pagesize, offset int) ([]any, error)

const PAGE_SIZE = 50

type AdapterField struct {
	Header     string
	Field      string
	CustomView func(row int, col int, dateItem interface{}) *tview.TableCell
}

type TViewTableAdapter struct {
	resolver   Resolver
	data       []map[string]interface{}
	withHeader bool
	fields     []*AdapterField
	lastOffset int
}

// Clear implements tview.TableContent.
func (*TViewTableAdapter) Clear() {
	panic("unimplemented")
}

// GetCell implements tview.TableContent.
func (a *TViewTableAdapter) GetCell(row int, column int) *tview.TableCell {
	dataIndex := row
	lastRow := row + 1
	if a.withHeader {
		dataIndex--
		lastRow = row
		if row == 0 {
			return tview.NewTableCell(a.fields[column].Header).SetSelectable(false)
		}
	}
	if len(a.data) == lastRow && lastRow >= a.lastOffset {
		requestedData, err := a.resolver(PAGE_SIZE, len(a.data))
		if err != nil {
			panic(err)
		}
		a.data = append(a.data, resolverWrapper(requestedData)...)
		a.lastOffset = len(a.data)
	}

	value := a.data[dataIndex][a.fields[column].Field]

	if a.fields[column].CustomView != nil {
		return a.fields[column].CustomView(row, column, value)
	} else {
		var cell *tview.TableCell
		valueType := reflect.TypeOf(value)
		switch valueType.Kind() {
		case reflect.String:
			cell = tview.NewTableCell(value.(string))
		case reflect.Int:
			cell = tview.NewTableCell(fmt.Sprintf("%d", value.(int)))
		case reflect.Bool:
			if value.(bool) {
				cell = tview.NewTableCell("Yes")
			} else {
				cell = tview.NewTableCell("No")
			}
		}
		cell.SetReference(value)
		return cell
	}

	// return tview.NewTableCell("")
}

// GetColumnCount implements tview.TableContent.
func (a *TViewTableAdapter) GetColumnCount() int {
	return len(a.fields)
}

// GetRowCount implements tview.TableContent.
func (a *TViewTableAdapter) GetRowCount() int {

	if len(a.data) == 0 {

		requestedData, err := a.resolver(PAGE_SIZE, 0)
		if err != nil {
			panic(err)
		}

		a.data = append(a.data, resolverWrapper(requestedData)...)
	}

	count := len(a.data)
	if a.withHeader {
		count++
	}
	return count
}

// InsertColumn implements tview.TableContent.
func (*TViewTableAdapter) InsertColumn(column int) {
	panic("unimplemented")
}

// InsertRow implements tview.TableContent.
func (*TViewTableAdapter) InsertRow(row int) {
	panic("unimplemented")
}

// RemoveColumn implements tview.TableContent.
func (*TViewTableAdapter) RemoveColumn(column int) {
	panic("unimplemented")
}

// RemoveRow implements tview.TableContent.
func (*TViewTableAdapter) RemoveRow(row int) {
	panic("unimplemented")
}

// SetCell implements tview.TableContent.
func (*TViewTableAdapter) SetCell(row int, column int, cell *tview.TableCell) {
	panic("unimplemented")
}

func InitReflectAdapter(resolver Resolver, header bool, fields ...*AdapterField) tview.TableContent {
	return &TViewTableAdapter{
		resolver:   resolver,
		fields:     fields,
		withHeader: header,
	}
}

func resolverWrapper(data []interface{}) []map[string]interface{} {
	var response []map[string]interface{}
	for _, v := range data {

		if reflect.TypeOf(v).Elem().Kind() == reflect.Struct {
			item := make(map[string]interface{})
			reflectedValue := reflect.ValueOf(v)

			for i := 0; i < reflectedValue.Elem().NumField(); i++ {
				field := reflectedValue.Elem().Field(i)
				fieldName := reflectedValue.Elem().Type().Field(i).Name
				item[fieldName] = field.Interface()
			}

			response = append(response, item)
		}
	}
	return response
}
