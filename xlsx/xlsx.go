package xlsx

import (
	"bytes"
	"fmt"
	"github.com/xuri/excelize/v2"
)

type ExcelFile struct {
	file      *excelize.File
	sheetName string
}

func NewExcelFile() *ExcelFile {
	sheetName := "Sheet1"
	file := excelize.NewFile()
	file.SetActiveSheet(file.NewSheet(sheetName))
	return &ExcelFile{file: file, sheetName: sheetName}
}

func (e *ExcelFile) FillExcelFile(params []string, data [][]interface{}) (*bytes.Buffer, error) {
	e.fillTop(params)
	e.fillData(data)
	return e.file.WriteToBuffer()
}

func (e *ExcelFile) fillTop(params []string) {
	topStyle, _ := e.file.NewStyle(`{"alignment":{"horizontal":"center","vertical":"center"}}`)
	word := 'A'
	for _, param := range params {
		line := fmt.Sprintf("%c1", word)
		_ = e.file.SetCellValue(e.sheetName, line, param)
		_ = e.file.SetCellStyle(e.sheetName, line, line, topStyle)
		word++
	}
}

func (e *ExcelFile) fillData(dataArr [][]interface{}) {
	lineStyle, _ := e.file.NewStyle(`{"alignment":{"horizontal":"center","vertical":"center"}}`)
	j := 2
	for _, arr := range dataArr {
		word := 'A'
		for _, data := range arr {
			line := fmt.Sprintf("%c%v", word, j)
			_ = e.file.SetCellValue(e.sheetName, line, data)
			_ = e.file.SetCellStyle(e.sheetName, line, line, lineStyle)
			word++
		}
		j++
	}
}
