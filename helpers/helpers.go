package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var (
	// ErrEmptyStringInToInt thrown when an empty string is passed into ToInt
	ErrEmptyStringInToInt = errors.New("Empty String passed to ToInt")
)

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
		return fmt.Errorf("Error outputting JSON: %s", err)
	}

	if string(output) == "null" {
		fmt.Println("[]")
	} else {
		fmt.Println(string(output))
	}

	return nil
}

func ToInt(value string) (int, error) {
	if value == "" {
		return 0, ErrEmptyStringInToInt
	}
	return strconv.Atoi(value)
}
