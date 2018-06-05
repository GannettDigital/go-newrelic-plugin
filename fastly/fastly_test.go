package fastly

import (
	"testing"

	fake "github.com/GannettDigital/paas-api-utils/utilsHTTP/fake"
	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

var fakeConfig Config

func init() {
	fakeConfig = Config{
		ServiceID:    "1234",
		FastlyAPIKey: "derp",
	}
}

func TestConvertToNrMetric(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		Stats           FastlyStats
	}{
		{
			TestDescription: "Adds expected stats from FastlyStats type to map[string]interface{}",
			Stats: FastlyStats{
				Requests:   1,
				HeaderSize: 2,
			},
		},
	}

	for _, test := range tests {
		g.Describe("convertToNrMetric()", func() {
			g.It(test.TestDescription, func() {
				result := convertToNrMetric(test.Stats, "derp", fakeConfig, logrus.New())
				g.Assert(result["fastly.datacenter"]).Equal("derp")
				g.Assert(result["fastly.requests"]).Equal(test.Stats.Requests)
			})
		})
	}
}

func TestGetFastlyStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		ExpectedLength  int
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/v1/channel/1234/ts/h",
						Code:   200,
						Data: []byte(`
							{
								"Data": [{
									"datacenter": {
										"HHN": {
											"requests": 14,
											"header_size": 13405,
											"body_size": 469254,
											"req_header_bytes": 6644,
											"resp_header_bytes": 13405,
											"resp_body_bytes": 469254,
											"bereq_header_bytes": 19628,
											"status_2xx": 4,
											"status_3xx": 10,
											"status_200": 4,
											"status_301": 1,
											"status_302": 9,
											"hits": 0,
											"miss": 2,
											"pass": 12,
											"synth": 0,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 3.669585,
											"miss_histogram": {
												"90": 1,
												"190": 2,
												"200": 1,
												"250": 2,
												"300": 1,
												"450": 1,
												"1600": 1,
												"2100": 1,
												"3500": 1,
												"4000": 1,
												"5000": 1,
												"7000": 1
											}
										},
										"MIA": {
											"requests": 17,
											"header_size": 8267,
											"body_size": 35417,
											"req_header_bytes": 58215,
											"resp_header_bytes": 8267,
											"resp_body_bytes": 35417,
											"bereq_header_bytes": 16913,
											"status_2xx": 4,
											"status_3xx": 13,
											"status_200": 4,
											"status_302": 3,
											"status_304": 10,
											"hits": 12,
											"miss": 2,
											"pass": 3,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.002911,
											"miss_time": 0.054782,
											"miss_histogram": {
												"20": 1,
												"30": 1,
												"190": 1,
												"350": 1,
												"1500": 1
											}
										},
										"LCY": {
											"requests": 3,
											"header_size": 2126,
											"body_size": 20510,
											"req_header_bytes": 2272,
											"resp_header_bytes": 2126,
											"resp_body_bytes": 20510,
											"bereq_header_bytes": 3664,
											"status_2xx": 1,
											"status_3xx": 2,
											"status_200": 1,
											"status_302": 2,
											"hits": 0,
											"miss": 1,
											"pass": 1,
											"synth": 1,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 0.571671,
											"miss_histogram": {
												"300": 1,
												"550": 1
											}
										},
										"JFK": {
											"requests": 233,
											"header_size": 152485,
											"body_size": 7262453,
											"req_header_bytes": 778609,
											"resp_header_bytes": 152485,
											"resp_body_bytes": 7262453,
											"bereq_header_bytes": 109265,
											"status_2xx": 114,
											"status_3xx": 116,
											"status_4xx": 3,
											"status_200": 64,
											"status_301": 6,
											"status_302": 43,
											"status_304": 67,
											"hits": 186,
											"miss": 14,
											"pass": 33,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.033215,
											"miss_time": 1.01477,
											"miss_histogram": {
												"8": 4,
												"9": 2,
												"10": 5,
												"20": 6,
												"30": 1,
												"60": 3,
												"90": 1,
												"100": 1,
												"170": 1,
												"180": 2,
												"190": 2,
												"200": 1,
												"250": 1,
												"550": 1,
												"650": 1,
												"950": 1,
												"1000": 1,
												"1100": 1,
												"1300": 1,
												"1500": 1,
												"1600": 1,
												"1900": 1,
												"2000": 1,
												"2300": 2,
												"2400": 1,
												"2600": 1,
												"3000": 1,
												"5500": 2
											}
										},
										"DFW": {
											"requests": 219,
											"header_size": 132951,
											"body_size": 3978768,
											"req_header_bytes": 584343,
											"resp_header_bytes": 132951,
											"resp_body_bytes": 3978768,
											"bereq_header_bytes": 133428,
											"tls": 7,
											"status_2xx": 90,
											"status_3xx": 124,
											"status_4xx": 5,
											"status_200": 73,
											"status_301": 6,
											"status_302": 54,
											"status_304": 64,
											"hits": 163,
											"miss": 29,
											"pass": 27,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.032923006,
											"miss_time": 87.60137,
											"miss_histogram": {
												"30": 5,
												"40": 6,
												"50": 1,
												"60": 2,
												"70": 3,
												"80": 3,
												"90": 1,
												"100": 1,
												"110": 1,
												"120": 2,
												"160": 1,
												"230": 2,
												"250": 1,
												"450": 1,
												"750": 1,
												"800": 1,
												"950": 1,
												"1100": 1,
												"1600": 1,
												"1700": 1,
												"1800": 1,
												"1900": 1,
												"2000": 1,
												"2100": 1,
												"2600": 1,
												"2900": 1,
												"3000": 2,
												"3500": 1,
												"4500": 2,
												"6500": 3,
												"7000": 2,
												"7500": 3,
												"10000": 1
											}
										},
										"BOS": {
											"requests": 11,
											"header_size": 8057,
											"body_size": 78364,
											"req_header_bytes": 26995,
											"resp_header_bytes": 8057,
											"resp_body_bytes": 78364,
											"bereq_header_bytes": 13290,
											"status_2xx": 2,
											"status_3xx": 8,
											"status_4xx": 1,
											"status_200": 2,
											"status_302": 4,
											"status_304": 4,
											"hits": 6,
											"miss": 1,
											"pass": 4,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.001375,
											"miss_time": 0.013598,
											"miss_histogram": {
												"10": 1,
												"30": 1,
												"40": 1,
												"60": 1,
												"170": 1
											}
										},
										"ATL": {
											"requests": 156,
											"header_size": 92468,
											"body_size": 1951369,
											"req_header_bytes": 564617,
											"resp_header_bytes": 92468,
											"resp_body_bytes": 1951369,
											"bereq_header_bytes": 180356,
											"tls": 2,
											"status_2xx": 57,
											"status_3xx": 97,
											"status_4xx": 2,
											"status_200": 55,
											"status_301": 4,
											"status_302": 35,
											"status_304": 58,
											"hits": 109,
											"miss": 20,
											"pass": 25,
											"synth": 2,
											"errors": 0,
											"hits_time": 0.030092003,
											"miss_time": 15.173494,
											"miss_histogram": {
												"10": 4,
												"20": 1,
												"30": 3,
												"50": 1,
												"70": 1,
												"80": 7,
												"90": 3,
												"100": 1,
												"110": 2,
												"150": 2,
												"190": 2,
												"250": 1,
												"300": 1,
												"350": 3,
												"500": 1,
												"1100": 1,
												"1200": 2,
												"1400": 1,
												"1500": 1,
												"1800": 1,
												"2300": 1,
												"2500": 1,
												"2800": 1,
												"3000": 1,
												"3500": 2
											}
										},
										"YYZ": {
											"requests": 38,
											"header_size": 19014,
											"body_size": 526016,
											"req_header_bytes": 175566,
											"resp_header_bytes": 19014,
											"resp_body_bytes": 526016,
											"bereq_header_bytes": 50150,
											"status_2xx": 10,
											"status_3xx": 28,
											"status_200": 9,
											"status_302": 11,
											"status_304": 17,
											"hits": 28,
											"miss": 10,
											"pass": 0,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.007008,
											"miss_time": 0.4554,
											"miss_histogram": {
												"20": 7,
												"80": 1,
												"90": 1,
												"110": 1
											}
										},
										"BMA": {
											"requests": 3,
											"header_size": 2196,
											"body_size": 0,
											"req_header_bytes": 652,
											"resp_header_bytes": 2196,
											"resp_body_bytes": 0,
											"bereq_header_bytes": 3353,
											"status_3xx": 3,
											"status_302": 3,
											"hits": 0,
											"miss": 0,
											"pass": 3,
											"synth": 0,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 0,
											"miss_histogram": {
												"350": 1,
												"450": 2
											}
										},
										"IAD": {
											"requests": 576,
											"header_size": 463763,
											"body_size": 7013701,
											"req_header_bytes": 1477553,
											"resp_header_bytes": 463763,
											"resp_body_bytes": 7013701,
											"bereq_header_bytes": 814988,
											"tls": 11,
											"shield": 198,
											"status_2xx": 265,
											"status_3xx": 298,
											"status_4xx": 12,
											"status_5xx": 1,
											"status_200": 265,
											"status_301": 54,
											"status_302": 165,
											"status_304": 79,
											"hits": 242,
											"miss": 164,
											"pass": 169,
											"synth": 1,
											"errors": 0,
											"hits_time": 0.045915995,
											"miss_time": 391.2639,
											"miss_histogram": {
												"7": 4,
												"8": 1,
												"9": 2,
												"10": 70,
												"20": 13,
												"30": 5,
												"40": 3,
												"50": 7,
												"60": 4,
												"70": 1,
												"80": 4,
												"90": 10,
												"100": 5,
												"110": 10,
												"120": 5,
												"130": 3,
												"140": 5,
												"150": 2,
												"160": 2,
												"170": 5,
												"180": 5,
												"190": 8,
												"210": 1,
												"230": 2,
												"240": 1,
												"250": 4,
												"300": 6,
												"350": 4,
												"400": 3,
												"450": 3,
												"500": 1,
												"550": 3,
												"600": 3,
												"700": 3,
												"750": 3,
												"850": 4,
												"950": 2,
												"1000": 4,
												"1100": 6,
												"1200": 3,
												"1400": 5,
												"1500": 3,
												"1600": 2,
												"1700": 2,
												"1800": 4,
												"1900": 1,
												"2000": 3,
												"2100": 2,
												"2200": 4,
												"2300": 2,
												"2400": 2,
												"2500": 2,
												"2600": 3,
												"2700": 3,
												"2800": 2,
												"3000": 10,
												"3500": 5,
												"4000": 4,
												"4500": 3,
												"5000": 4,
												"5500": 9,
												"6000": 4,
												"6500": 7,
												"7000": 7,
												"7500": 6,
												"9500": 1,
												"10000": 2,
												"12000": 1
											}
										},
										"SJC": {
											"requests": 37,
											"header_size": 24957,
											"body_size": 343342,
											"req_header_bytes": 100043,
											"resp_header_bytes": 24957,
											"resp_body_bytes": 343342,
											"bereq_header_bytes": 71686,
											"status_2xx": 16,
											"status_3xx": 21,
											"status_200": 16,
											"status_302": 16,
											"status_304": 5,
											"hits": 16,
											"miss": 9,
											"pass": 12,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.004149,
											"miss_time": 6.1543617,
											"miss_histogram": {
												"70": 3,
												"80": 4,
												"90": 1,
												"100": 1,
												"130": 1,
												"140": 2,
												"150": 2,
												"250": 1,
												"300": 1,
												"950": 1,
												"2100": 1,
												"2200": 1,
												"6000": 1,
												"8000": 1
											}
										},
										"MSP": {
											"requests": 54,
											"header_size": 33504,
											"body_size": 880753,
											"req_header_bytes": 178746,
											"resp_header_bytes": 33504,
											"resp_body_bytes": 880753,
											"bereq_header_bytes": 36828,
											"status_2xx": 26,
											"status_3xx": 27,
											"status_4xx": 1,
											"status_200": 20,
											"status_301": 2,
											"status_302": 5,
											"status_304": 20,
											"hits": 44,
											"miss": 3,
											"pass": 7,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.009957,
											"miss_time": 0.45979398,
											"miss_histogram": {
												"30": 1,
												"40": 1,
												"50": 2,
												"160": 1,
												"170": 1,
												"240": 1,
												"250": 2,
												"17000": 1
											}
										},
										"DEN": {
											"requests": 16,
											"header_size": 8817,
											"body_size": 26478,
											"req_header_bytes": 46271,
											"resp_header_bytes": 8817,
											"resp_body_bytes": 26478,
											"bereq_header_bytes": 23136,
											"status_2xx": 3,
											"status_3xx": 13,
											"status_200": 3,
											"status_301": 2,
											"status_302": 6,
											"status_304": 5,
											"hits": 9,
											"miss": 3,
											"pass": 4,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.002075,
											"miss_time": 0.203232,
											"miss_histogram": {
												"40": 1,
												"60": 1,
												"100": 1,
												"110": 1,
												"140": 1,
												"450": 1,
												"1600": 1
											}
										},
										"SYD": {
											"requests": 1,
											"header_size": 853,
											"body_size": 0,
											"req_header_bytes": 260,
											"resp_header_bytes": 853,
											"resp_body_bytes": 0,
											"bereq_header_bytes": 1098,
											"status_3xx": 1,
											"status_302": 1,
											"hits": 0,
											"miss": 0,
											"pass": 1,
											"synth": 0,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 0,
											"miss_histogram": {
												"400": 1
											}
										},
										"LAX": {
											"requests": 61,
											"header_size": 38182,
											"body_size": 211056,
											"req_header_bytes": 185835,
											"resp_header_bytes": 38182,
											"resp_body_bytes": 211056,
											"bereq_header_bytes": 127582,
											"status_2xx": 21,
											"status_3xx": 35,
											"status_4xx": 4,
											"status_5xx": 1,
											"status_200": 21,
											"status_301": 5,
											"status_302": 22,
											"status_304": 8,
											"hits": 28,
											"miss": 13,
											"pass": 19,
											"synth": 1,
											"errors": 0,
											"hits_time": 0.007074001,
											"miss_time": 13.120308,
											"miss_histogram": {
												"30": 1,
												"40": 2,
												"60": 4,
												"70": 6,
												"80": 1,
												"150": 2,
												"250": 2,
												"500": 1,
												"900": 1,
												"1000": 1,
												"1100": 1,
												"1300": 1,
												"1600": 1,
												"1700": 1,
												"1800": 2,
												"2000": 3,
												"2500": 1,
												"3500": 1
											}
										},
										"AMS": {
											"requests": 11,
											"header_size": 10733,
											"body_size": 132984,
											"req_header_bytes": 4871,
											"resp_header_bytes": 10733,
											"resp_body_bytes": 132984,
											"bereq_header_bytes": 13900,
											"status_2xx": 2,
											"status_3xx": 9,
											"status_200": 2,
											"status_302": 9,
											"hits": 1,
											"miss": 1,
											"pass": 9,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.000244,
											"miss_time": 0.088364,
											"miss_histogram": {
												"80": 1,
												"100": 1,
												"170": 3,
												"250": 1,
												"300": 1,
												"1800": 1,
												"1900": 1,
												"2800": 1
											}
										},
										"SIN": {
											"requests": 2,
											"header_size": 1446,
											"body_size": 13695,
											"req_header_bytes": 5220,
											"resp_header_bytes": 1446,
											"resp_body_bytes": 13695,
											"bereq_header_bytes": 7769,
											"status_2xx": 1,
											"status_3xx": 1,
											"status_200": 1,
											"status_302": 1,
											"hits": 0,
											"miss": 1,
											"pass": 1,
											"synth": 0,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 0.24973,
											"miss_histogram": {
												"250": 2
											}
										},
										"NRT": {
											"requests": 1,
											"header_size": 765,
											"body_size": 0,
											"req_header_bytes": 292,
											"resp_header_bytes": 765,
											"resp_body_bytes": 0,
											"bereq_header_bytes": 1079,
											"status_3xx": 1,
											"status_302": 1,
											"hits": 0,
											"miss": 0,
											"pass": 1,
											"synth": 0,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 0,
											"miss_histogram": {
												"180": 1
											}
										},
										"SEA": {
											"requests": 106,
											"header_size": 71275,
											"body_size": 2117020,
											"req_header_bytes": 121347,
											"resp_header_bytes": 71275,
											"resp_body_bytes": 2117020,
											"bereq_header_bytes": 104855,
											"status_2xx": 63,
											"status_3xx": 42,
											"status_4xx": 1,
											"status_200": 63,
											"status_301": 13,
											"status_302": 15,
											"status_304": 14,
											"hits": 42,
											"miss": 54,
											"pass": 10,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.0057059997,
											"miss_time": 176.95245,
											"miss_histogram": {
												"20": 2,
												"30": 1,
												"40": 1,
												"70": 1,
												"80": 2,
												"90": 1,
												"110": 1,
												"120": 1,
												"130": 1,
												"150": 1,
												"170": 1,
												"210": 1,
												"250": 1,
												"300": 1,
												"350": 1,
												"400": 1,
												"450": 1,
												"500": 2,
												"600": 2,
												"800": 1,
												"900": 1,
												"950": 1,
												"1000": 2,
												"1200": 1,
												"1600": 1,
												"1700": 1,
												"1800": 1,
												"1900": 2,
												"2400": 2,
												"2500": 1,
												"2700": 1,
												"3000": 1,
												"3500": 1,
												"4000": 4,
												"4500": 2,
												"5000": 2,
												"5500": 5,
												"6000": 4,
												"7000": 4,
												"7500": 2,
												"10000": 1
											}
										},
										"ORD": {
											"requests": 419,
											"header_size": 253402,
											"body_size": 2959460,
											"req_header_bytes": 1375986,
											"resp_header_bytes": 253402,
											"resp_body_bytes": 2959460,
											"bereq_header_bytes": 344168,
											"tls": 7,
											"http2": 2,
											"status_2xx": 152,
											"status_3xx": 258,
											"status_4xx": 9,
											"status_200": 152,
											"status_301": 10,
											"status_302": 123,
											"status_304": 125,
											"hits": 306,
											"miss": 30,
											"pass": 83,
											"synth": 0,
											"errors": 0,
											"hits_time": 0.085706,
											"miss_time": 22.095474,
											"miss_histogram": {
												"20": 9,
												"30": 17,
												"40": 10,
												"50": 5,
												"60": 7,
												"70": 4,
												"80": 2,
												"90": 1,
												"100": 4,
												"110": 6,
												"120": 1,
												"130": 6,
												"150": 2,
												"170": 1,
												"190": 1,
												"210": 1,
												"240": 1,
												"250": 2,
												"300": 2,
												"550": 1,
												"600": 1,
												"650": 1,
												"700": 1,
												"750": 2,
												"850": 2,
												"950": 2,
												"1000": 3,
												"1300": 1,
												"1400": 1,
												"1500": 1,
												"2000": 1,
												"2300": 1,
												"2500": 2,
												"2700": 3,
												"2900": 1,
												"3000": 1,
												"3500": 1,
												"6000": 1,
												"7500": 1,
												"8500": 1,
												"10000": 1,
												"15000": 1
											}
										},
										"FRA": {
											"requests": 6,
											"header_size": 4908,
											"body_size": 139062,
											"req_header_bytes": 2893,
											"resp_header_bytes": 4908,
											"resp_body_bytes": 139062,
											"bereq_header_bytes": 5718,
											"status_2xx": 2,
											"status_3xx": 4,
											"status_200": 2,
											"status_301": 2,
											"status_302": 2,
											"hits": 2,
											"miss": 0,
											"pass": 3,
											"synth": 1,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 0,
											"miss_histogram": {
												"220": 1,
												"600": 1,
												"3000": 1
											}
										},
										"LHR": {
											"requests": 6,
											"header_size": 5162,
											"body_size": 44195,
											"req_header_bytes": 2732,
											"resp_header_bytes": 5162,
											"resp_body_bytes": 44195,
											"bereq_header_bytes": 6293,
											"status_2xx": 2,
											"status_3xx": 4,
											"status_200": 2,
											"status_301": 2,
											"status_302": 2,
											"hits": 1,
											"miss": 1,
											"pass": 4,
											"synth": 0,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 1.575046,
											"miss_histogram": {
												"80": 1,
												"160": 1,
												"300": 1,
												"1500": 1,
												"5000": 1
											}
										},
										"HKG": {
											"requests": 3,
											"header_size": 3085,
											"body_size": 46849,
											"req_header_bytes": 1024,
											"resp_header_bytes": 3085,
											"resp_body_bytes": 46849,
											"bereq_header_bytes": 3619,
											"status_2xx": 1,
											"status_3xx": 1,
											"status_4xx": 1,
											"status_200": 1,
											"status_302": 1,
											"hits": 0,
											"miss": 0,
											"pass": 3,
											"synth": 0,
											"errors": 0,
											"hits_time": 0,
											"miss_time": 0,
											"miss_histogram": {
												"400": 1,
												"1900": 1,
												"5500": 1
											}
										}
									},
									"aggregated": {
										"requests": 1993,
										"header_size": 1351821,
										"body_size": 28250746,
										"req_header_bytes": 5700986,
										"resp_header_bytes": 1351821,
										"resp_body_bytes": 28250746,
										"bereq_header_bytes": 2092766,
										"tls": 27,
										"shield": 198,
										"http2": 2,
										"status_2xx": 836,
										"status_3xx": 1116,
										"status_4xx": 39,
										"status_5xx": 2,
										"status_200": 760,
										"status_301": 107,
										"status_302": 533,
										"status_304": 476,
										"hits": 1195,
										"miss": 358,
										"pass": 434,
										"synth": 6,
										"errors": 0,
										"hits_time": 0.26835108,
										"miss_time": 720.71747,
										"miss_histogram": {
											"7": 4,
											"8": 5,
											"9": 4,
											"10": 80,
											"20": 39,
											"30": 36,
											"40": 25,
											"50": 16,
											"60": 22,
											"70": 19,
											"80": 26,
											"90": 20,
											"100": 15,
											"110": 22,
											"120": 9,
											"130": 11,
											"140": 8,
											"150": 11,
											"160": 5,
											"170": 13,
											"180": 8,
											"190": 16,
											"200": 2,
											"210": 3,
											"220": 1,
											"230": 4,
											"240": 3,
											"250": 20,
											"300": 15,
											"350": 10,
											"400": 6,
											"450": 9,
											"500": 5,
											"550": 6,
											"600": 7,
											"650": 2,
											"700": 4,
											"750": 6,
											"800": 2,
											"850": 6,
											"900": 2,
											"950": 8,
											"1000": 11,
											"1100": 10,
											"1200": 6,
											"1300": 3,
											"1400": 7,
											"1500": 8,
											"1600": 8,
											"1700": 5,
											"1800": 10,
											"1900": 7,
											"2000": 9,
											"2100": 5,
											"2200": 5,
											"2300": 6,
											"2400": 5,
											"2500": 7,
											"2600": 5,
											"2700": 7,
											"2800": 4,
											"2900": 2,
											"3000": 17,
											"3500": 12,
											"4000": 9,
											"4500": 7,
											"5000": 8,
											"5500": 17,
											"6000": 10,
											"6500": 10,
											"7000": 14,
											"7500": 12,
											"8000": 1,
											"8500": 1,
											"9500": 1,
											"10000": 5,
											"12000": 1,
											"15000": 1,
											"17000": 1
										}
									},
									"recorded": 1497276984
								}],
								"Timestamp": 1497276993,
								"AggregateDelay": 9
							}
							`),
						Err: nil,
					},
				},
			},
			ExpectedLength:  1,
			TestDescription: "Successfully GET fastly stats",
		},
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/v1/channel/1234/ts/h",
						Code:   500,
						Data:   []byte(``),
						Err:    nil,
					},
				},
			},
			ExpectedLength:  0,
			TestDescription: "Successfully not pannic when a non 200 result occurrs",
		},
	}

	for _, test := range tests {
		g.Describe("getFastlyStats()", func() {
			g.It(test.TestDescription, func() {
				runner = &test.HTTPRunner
				result := getFastlyStats(logrus.New(), fakeConfig)
				g.Assert(len(result.Data)).Equal(test.ExpectedLength)
				if len(result.Data) > 0 {
					g.Assert(result.Data[0].Aggregated.Hits).Equal(1195)
				}
			})
		})
	}
}
