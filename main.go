package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type iomode int

const (
	unknown iomode = iota - 1
	yaml2csv
	csv2yaml
)

var (
	errUnknownMode  = errors.New("unknown mode")
	errUnknownError = errors.New("unknown error")
)

func main() {
	inputFormat := flag.String("m", "", "csv or yaml")
	flag.Parse()

	if err := mainWithErr(*inputFormat); err != nil {
		io.WriteString(os.Stderr, fmt.Sprintf("error: %s\n", err.Error()))
	}
}

func mainWithErr(inputFormat string) error {
	var (
		err   error
		input []byte
		mode  iomode
	)

	if input, err = slurpInput(); err != nil {
		return err
	}

	if inputFormat != "" {
		switch inputFormat {
		case "yaml":
			mode = yaml2csv
		case "csv":
			mode = csv2yaml
		default:
			return fmt.Errorf("%w: %s", errUnknownMode, inputFormat)
		}
	} else {
		mode, err = sniffIOMode(input)
		if err != nil {
			return err
		}
	}

	switch mode {
	case yaml2csv:
		return convertYAML2CSV(input)
	case csv2yaml:
		return convertCSV2YAML(input)
	}

	return errUnknownError
}

func slurpInput() ([]byte, error) {
	var (
		err   error
		input []byte
	)

	if input, err = ioutil.ReadAll(os.Stdin); err != nil {
		return nil, err
	}

	return input, nil
}

var (
	looksLikeYAML = regexp.MustCompile(`^(---$|- [^:]*:)`)
)

func sniffIOMode(input []byte) (iomode, error) {
	r := bufio.NewReader(bytes.NewReader(input))

	firstLine, err := r.ReadString('\n')
	if err != nil || firstLine == "" {
		return unknown, err
	}

	matchesYAML := looksLikeYAML.MatchString(strings.TrimSpace(firstLine))

	if matchesYAML {
		return yaml2csv, nil
	}

	return csv2yaml, nil
}

func convertYAML2CSV(input []byte) error {
	var (
		err     error
		yamlDoc []map[string]string
		csvCols []string
		csvData [][]string
	)

	if err = yaml.Unmarshal(input, &yamlDoc); err != nil {
		return err
	}

	if len(yamlDoc) == 0 {
		io.WriteString(os.Stderr, "warning: no input records\n")
	}

	// Check if the first record is a __meta__ record
	if _, ok := yamlDoc[0]["__meta__"]; ok {
		csvCols = strings.Split(yamlDoc[0]["column_order"], ",")
		yamlDoc = yamlDoc[1:]
	} else {
		// Otherwise use the keys from the first record in whichever order they're iterated
		for k := range yamlDoc[0] {
			csvCols = append(csvCols, k)
		}
	}
	csvData = append(csvData, csvCols)

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

func convertCSV2YAML(input []byte) error {
	var (
		err      error
		csvData  [][]string
		yamlData []map[string]string
	)

	r := csv.NewReader(bytes.NewReader(input))

	if csvData, err = r.ReadAll(); err != nil {
		return err
	}

	if len(csvData) < 2 {
		io.WriteString(os.Stderr, "warning: no input records\n")
	}

	csvCols := csvData[0]

	// create the __meta__ record
	yamlData = append(yamlData, map[string]string{
		"__meta__":     "",
		"column_order": strings.Join(csvCols, ","),
	})

	for _, r := range csvData[1:] {
		yamlRec := map[string]string{}
		for i, v := range r {
			yamlRec[csvCols[i]] = v
		}
		yamlData = append(yamlData, yamlRec)
	}

	var yamlOut []byte
	if yamlOut, err = yaml.Marshal(yamlData); err != nil {
		return err
	}

	if _, err = io.WriteString(os.Stdout, string(yamlOut)); err != nil {
		return err
	}

	return nil
}
