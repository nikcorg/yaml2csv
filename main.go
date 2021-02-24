package main

import (
	"encoding/csv"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {
	if err := mainWithErr(); err != nil {
		io.WriteString(os.Stderr, err.Error())
	}
}

func mainWithErr() error {
	var (
		err      error
		yamlData []byte
		yamlDoc  []map[string]string
		csvCols  []string
		csvData  [][]string
	)

	if yamlData, err = ioutil.ReadAll(os.Stdin); err != nil {
		return err
	}

	if err = yaml.Unmarshal(yamlData, &yamlDoc); err != nil {
		return err
	}

	if len(yamlDoc) == 0 {
		return errors.New("no input records")
	}

	for k := range yamlDoc[0] {
		csvCols = append(csvCols, k)
	}
	csvData = append(csvData, csvCols)

	// Append records
	for _, r := range yamlDoc {
		var record []string
		for _, k := range csvCols {
			record = append(record, r[k])
		}
		csvData = append(csvData, record)
	}

	if err = csv.NewWriter(os.Stdout).WriteAll(csvData); err != nil {
		return err
	}

	return nil
}
