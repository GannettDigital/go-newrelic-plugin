package jira

import (
	"testing"
)

func TestGetOpenIssues(t *testing.T) {

}

// func TestGetIssues(t *testing.T) {
// 	g := goblin.Goblin(t)
// 	j, _ := NewJira(Config{authToken: "somefakeauth"})

// 	var tests = []struct {
// 		HTTPRunner      fake.HTTPResult
// 		TestDescription string
// 	}{
// 		{
// 			HTTPRunner: fake.HTTPResult{
// 				ResultsList: []fake.Result{
// 					{
// 						Method: "GET",
// 						URI:    "/rest/api/2/search",
// 						Code:   200,
// 						Data: []byte(`
// 							[{
// 								"issues": "something
// 							}]`),
// 						Err: nil,
// 					},
// 				},
// 			},
// 			TestDescription: "Successfully GET jira issues",
// 		},
// 	}

// 	for _, test := range tests {
// 		g.Describe("GetJiraOpenIssues", func() {
// 			g.It(test.TestDescription, func() {
// 				runner := &test.HTTPRunner
// 				result, err := j.GetOpenIssues(runner)
// 				g.Assert(err).Equal(nil)
// 				g.Assert(result).Equal("something")
// 			})
// 		})
// 	}
// }
