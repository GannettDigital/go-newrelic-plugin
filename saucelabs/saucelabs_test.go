package saucelabs

import (
	"os"
	"testing"

	fake "github.com/GannettDigital/paas-api-utils/utilsHTTP/fake"
	"github.com/franela/goblin"
)

var fakeConfig SauceConfig

func init() {
	fakeConfig = SauceConfig{
		SauceAPIUser: "test-user",
		SauceAPIKey:  "test-key",
	}
}

func TestUserActivity(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "rest/v1/" + os.Getenv("SAUCE_API_USER") + "/activity",
						Code:   200,
						Data: []byte(`{"subaccounts":{"steelers":{"in progress":7,"all":7,"queued":0},"penguins":{"in progress":1,"all":1,"queued":0},"pirates":{"in progress":1,"all":2,"queued":1}},"totals":{"in progress":9,"all":10,"queued":1}}`),
						Err: nil,
					},
				},
			},
			TestDescription: "Successfully GET saucelabs userActivity",
		},
	}
	for _, test := range tests {
		g.Describe("getUserActivity()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getUserActivity(client, fakeConfig)
				g.Assert(result.SubAccounts["steelers"].All).Equal(7)
			})
		})
	}
}

func TestGetConcurrency(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "rest/v1.1/" + os.Getenv("SAUCE_API_USER") + "/concurrency",
						Code:   200,
						Data: []byte(`{"timestamp":8.7712928E8,"concurrency":{"self":{"username":"testing","current":{"manual":0,"mac":1,"overall":4},"allowed":{"manual":43,"mac":43,"overall":43}},"ancestor":{"username":"testing","current":{"manual":0,"mac":1,"overall":4},"allowed":{"manual":43,"mac":43,"overall":43}}}}`),
						Err: nil,
					},
				},
			},
			TestDescription: "Successfully GET saucelabs concurrency",
		},
	}
	for _, test := range tests {
		g.Describe("getConcurrency()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getConcurrency(client, fakeConfig)
				g.Assert(result.TeamData["self"].Current.Overall).Equal(4)
			})
		})
	}
}
func TestGetUserList(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "rest/v1/" + os.Getenv("SAUCE_API_USER") + "/subaccounts",
						Code:   200,
						Data: []byte(`{"users_total":2,"users":[{"username":"FIRSTUSER","vm_lockdown":false,"new_email":null,"last_name":"Doe","parent":"gd_automation","subaccount_limit":null,"children_count":0,"creation_time":1465484232,"user_type":"subaccount","monthly_minutes":{"manual":"infinite","automated":999999},"prevent_emails":[],"is_admin":null,"manual_minutes":"infinite","can_run_manual":true,"concurrency_limit":{"mac":70,"scout":3,"overall":3,"real_device":0},"is_public":false,"id":"FIRSTUSER","access_key":"12345","first_name":"John","require_full_name":true,"verified":false,"name":"John Doe","subscribed":true,"title":null,"terminating_subscription":false,"is_sso":false,"entity_type":null,"tunnel_concurrency_limit":70,"allow_integrations_page":true,"last_login":null,"ancestor_concurrency_limit":{"mac":70,"scout":70,"overall":70,"real_device":0},"ancestor_allows_subaccounts":false,"domain":null,"ancestor":"gd_automation","minutes":980420,"email":"johndoe@gannett.com"},{"username":"seconduser","vm_lockdown":false,"new_email":null,"last_name":null,"parent":"gd_automation","subaccount_limit":null,"last_test":1502724106,"children_count":0,"creation_time":1502719650,"user_type":"subaccount","monthly_minutes":{"manual":"infinite","automated":999999},"prevent_emails":[],"is_admin":null,"manual_minutes":"infinite","can_run_manual":true,"concurrency_limit":{"mac":70,"scout":3,"overall":3,"real_device":0},"is_public":false,"id":"seconduser","access_key":"123743","first_name":null,"require_full_name":true,"verified":true,"name":"Jane Doe","subscribed":true,"title":null,"terminating_subscription":false,"is_sso":false,"entity_type":null,"tunnel_concurrency_limit":70,"allow_integrations_page":true,"last_login":1502722851,"ancestor_concurrency_limit":{"mac":70,"scout":70,"overall":70,"real_device":0},"ancestor_allows_subaccounts":false,"domain":null,"ancestor":"gd_automation","minutes":980420,"email":"janedoe@gannett.com"}]}`),
						Err: nil,
					},
				},
			},
			TestDescription: "Successfully GET saucelabs userlist",
		},
	}
	for _, test := range tests {
		g.Describe("getUserList()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getUserList(client, fakeConfig)
				g.Assert(result.users[0].UserName).Equal("FIRSTUSER")
			})
		})
	}
}

func TestGetUsage(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "rest/v1.1/" + os.Getenv("SAUCE_API_USER") + "/usage",
						Code:   200,
						Data: []byte(`{"usage":[["2017-7-14",[24,6509]],["2017-7-19",[2,266]],["2017-7-20",[6,591]],["2017-7-25",[2,145]],["2017-7-26",[20,7406]],["2017-8-4",[16,2076]],["2017-8-8",[8,702]],["2017-8-9",[24,8315]],["2017-8-10",[5,761]],["2017-8-11",[43,4116]],["2017-8-14",[529,19511]],["2017-8-15",[130,4423]],["2017-8-24",[21,528]],["2017-9-6",[86,5388]]],"username":"testing"}`),
						Err: nil,
					},
				},
			},
			TestDescription: "Successfully GET saucelabs usage",
		},
	}
	for _, test := range tests {
		g.Describe("getUsage()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getUsage(client, fakeConfig)
				g.Assert(result.usage[0][0]).Equal("2017-7-14")
			})
		})
	}
}
