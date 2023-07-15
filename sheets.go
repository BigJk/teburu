package teburu

import (
	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
)

var sheetFields = []googleapi.Field{
	"sheets/data/rowData/values/hyperlink",
	"sheets/data/rowData/values/effectiveValue",
}

type CellType string

const (
	CellTypeSimpleOnly = CellType("simple")
	CellTypeComplex    = CellType("complex")
	CellTypeDynamic    = CellType("dynamic")
)

type ComplexCellValue struct {
	Value interface{} `json:"value"`
	Link  string      `json:"link"`
}

// CollapseCell takes an effective value and a hyperlink and returns a value that can be marshalled to JSON.
func CollapseCell(effectiveValue *sheets.ExtendedValue, hyperlink string, cellType CellType) interface{} {
	if effectiveValue == nil {
		return ""
	}

	var value interface{}
	if effectiveValue.BoolValue != nil {
		value = *effectiveValue.BoolValue
	}
	if effectiveValue.NumberValue != nil {
		value = *effectiveValue.NumberValue
	}
	if effectiveValue.StringValue != nil {
		value = *effectiveValue.StringValue
	}
	if effectiveValue.FormulaValue != nil {
		value = *effectiveValue.FormulaValue
	}
	if effectiveValue.ErrorValue != nil {
		value = *effectiveValue.ErrorValue
	}

	if cellType == CellTypeSimpleOnly {
		return value
	} else if cellType == CellTypeDynamic {
		if hyperlink != "" {
			return ComplexCellValue{
				Value: value,
				Link:  hyperlink,
			}
		}
	} else if cellType == CellTypeComplex {
		return ComplexCellValue{
			Value: value,
			Link:  hyperlink,
		}
	}

	return value
}
