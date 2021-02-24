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
	errNoInputRecords    = errors.New("no input records")
	errUnknownMode       = errors.New("unknown mode")
	errUnknownError      = errors.New("unknown error")
	errUnsuccessfulSniff = errors.New("unrecognisable first line, set input format using -m (csv|yaml)")
)

var inputFormat string

func init() {
	flag.StringVar(&inputFormat, "m", "", "csv or yaml")
}

func main() {
	flag.Parse()
	if err := mainWithErr(); err != nil {
		io.WriteString(os.Stderr, fmt.Sprintf("error: %s\n", err.Error()))
	}
}

func mainWithErr() error {
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

var csvRe = regexp.MustCompile(`^[^,]+((,[^,]+)*)?$`)
var yamlRe = regexp.MustCompile(`^(---|- \w+:)$`)

func sniffIOMode(input []byte) (iomode, error) {
	r := bufio.NewReader(bytes.NewReader(input))

	firstLine, err := r.ReadString('\n')
	if err != nil {
		return unknown, err
	}

	firstLine = strings.TrimSpace(firstLine)
	matchesCSV := csvRe.MatchString(firstLine)
	matchesYAML := yamlRe.MatchString(firstLine)

	if !matchesCSV && !matchesYAML {
		return unknown, errUnsuccessfulSniff
	}

	fmt.Println(firstLine, matchesYAML, matchesCSV)

	switch {
	case matchesYAML:
		return yaml2csv, nil
	case matchesCSV:
		return csv2yaml, nil
	}

	return unknown, errUnknownError
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
		return errNoInputRecords
	}

	for k := range yamlDoc[0] {
		csvCols = append(csvCols, k)
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
		return errNoInputRecords
	}

	csvCols := csvData[0]

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
