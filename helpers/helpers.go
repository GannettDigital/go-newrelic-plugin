package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

// OutputJSON takes an object and prints it as a JSON string to the stdout.
// If the pretty attribute is set to true, the JSON will be idented for easy reading.
func OutputJSONLog(log *logrus.Logger, data interface{}, pretty bool) error {
	log.Info("3")
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "\t")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		log.Info("4")
		return fmt.Errorf("Error outputting JSON: %v \n data: %s", data, err)
	}

	if string(output) == "null" {
		log.Info("5")
		fmt.Println("[]")
	} else {
		log.Info("6")
		fmt.Println(string(output))
	}

	return nil
}

// OutputJSON takes an object and prints it as a JSON string to the stdout.
// If the pretty attribute is set to true, the JSON will be idented for easy reading.
func OutputJSON(data interface{}, pretty bool) error {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "\t")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("Error outputting JSON: %v \n data: %s", data, err)
	}

	if string(output) == "null" {
		fmt.Println("[]")
	} else {
		fmt.Println(string(output))
	}

	return nil
}

func CamelCase(src string) string {
	var camelingRegex = regexp.MustCompile("[0-9A-Za-z.]+")
	src = strings.Replace(src, ":", ".", -1)
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = bytes.Title(val)
		}
	}
	result := string(bytes.Join(chunks, nil))
	return result
}

func AsValue(value string) interface{} {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return value
}
