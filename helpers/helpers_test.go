package helpers

import (
	"testing"

	"github.com/franela/goblin"
)

func TestCamelCase(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		src             string
		result          string
	}{
		{
			TestDescription: "Should convert to camel case without error",
			src:             "one_two_three_four",
			result:          "oneTwoThreeFour",
		},
		{
			TestDescription: "Should convert to camel case with colon without error",
			src:             "one:two_three_four",
			result:          "one.twoThreeFour",
		},
	}
	for _, test := range tests {
		g.Describe("camelCase)", func() {
			g.It(test.TestDescription, func() {
				camel := CamelCase(test.src)
				g.Assert(camel).Equal(test.result)
			})
		})
	}
}

func TestOutputJSON(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputData       interface{}
		InputPretty     bool
		ExpectedErr     error
		TestDescription string
	}{
		{
			InputData: map[string]interface{}{
				"thing": "stuff",
			},
			InputPretty:     false,
			ExpectedErr:     nil,
			TestDescription: "Should return no error with valid input and pretty of false",
		},
		{
			InputData: map[string]interface{}{
				"thing": "stuff",
			},
			InputPretty:     true,
			ExpectedErr:     nil,
			TestDescription: "Should return no error with valid input and pretty of true",
		},
		{
			InputData:       nil,
			InputPretty:     false,
			ExpectedErr:     nil,
			TestDescription: "Should return no error when nil value is provided",
		},
	}

	for _, test := range tests {
		g.Describe("OutputJSON()", func() {
			g.It(test.TestDescription, func() {
				err := OutputJSON(test.InputData, test.InputPretty)
				g.Assert(err).Equal(test.ExpectedErr)
			})
		})
	}
}

func TestAsValue(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		value           string
		result          interface{}
	}{
		{
			TestDescription: "Should convert string to string without error",
			value:           "string",
			result:          "string",
		},
		{
			TestDescription: "Should convert string to int without error",
			value:           "1",
			result:          1,
		},
		{
			TestDescription: "Should convert string to float without error",
			value:           "1.0",
			result:          1.0,
		},
		{
			TestDescription: "Should convert string to boolean without error",
			value:           "true",
			result:          true,
		},
	}
	for _, test := range tests {
		g.Describe("asValue)", func() {
			g.It(test.TestDescription, func() {
				value := AsValue(test.value)
				g.Assert(value).Equal(test.result)
			})
		})
	}
}
