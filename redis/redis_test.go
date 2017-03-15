package redis

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

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

func TestInitRedisClient(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputConfig     Config
		TestDescription string
	}{
		{
			InputConfig: Config{
				RedisHost: "localhost",
				RedisPort: "6379",
				RedisPass: "",
				RedisDB:   "0",
			},
			TestDescription: "Should successfully create a redis client",
		},
	}

	for _, test := range tests {
		g.Describe("initRedisClient()", func() {
			g.It(test.TestDescription, func() {
				initRedisClient(test.InputConfig)
				g.Assert(true).IsTrue()
			})
		})
	}
}

func TestFatalIfErrt(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputErr        error
		TestDescription string
	}{
		{
			InputLog:        logrus.New(),
			InputErr:        nil,
			TestDescription: "Should successfully not exit on a nil error",
		},
	}

	for _, test := range tests {
		g.Describe("fatalIfErr()", func() {
			g.It(test.TestDescription, func() {
				fatalIfErr(test.InputLog, test.InputErr)
				g.Assert(true).IsTrue()
			})
		})
	}
}

func TestValidateConfig(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputConfig     Config
		ExpectedConfig  Config
		TestDescription string
	}{
		{
			InputLog:    logrus.New(),
			InputConfig: Config{},
			ExpectedConfig: Config{
				RedisHost: "localhost",
				RedisPort: "6379",
			},
			TestDescription: "Should successfully set proper defaults when none are provided",
		},
		{
			InputLog: logrus.New(),
			InputConfig: Config{
				RedisHost: "10.0.0.1",
				RedisPort: "1234",
				RedisPass: "somepass",
				RedisDB:   "2",
			},
			ExpectedConfig: Config{
				RedisHost: "10.0.0.1",
				RedisPort: "1234",
				RedisPass: "somepass",
				RedisDB:   "2",
				DBID:      2,
			},
			TestDescription: "Should successfully set proper defaults when none are provided",
		},
	}

	for _, test := range tests {
		g.Describe("validateConfig()", func() {
			g.It(test.TestDescription, func() {
				validateConfig(test.InputLog, &test.InputConfig)
				g.Assert(test.InputConfig).Equal(test.ExpectedConfig)
			})
		})
	}
}
