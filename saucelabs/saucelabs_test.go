package saucelabs

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
)

var (
	fakeLog    = logrus.New()
	fakeConfig = SauceConfig{
		SauceAPIUser: "test-user",
		SauceAPIKey:  "test-pw",
	}
)

// TestValidateConfig tests if the sauceconfig validator is working as expected
func TestValidateConfig(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("jenkins validateConfig()", func() {
		expected := map[string]struct {
			ExpectedIsNil bool
			SauceConfig   SauceConfig
		}{
			"no": {
				false,
				SauceConfig{},
			},
			"SauceAPIUser": {
				false,
				SauceConfig{
					SauceAPIUser: "test-user",
				},
			},
			"SauceAPIKey": {
				false,
				SauceConfig{
					SauceAPIKey: "test-pw",
				},
			},
			"all": {
				true,
				SauceConfig{SauceAPIUser: "test-user",
					SauceAPIKey: "test-pw",
				},
			},
		}
		for name, ex := range expected {
			desc := fmt.Sprintf("should return %v when %v fields are set", ex.ExpectedIsNil, name)
			g.It(desc, func() {
				valid := validateConfig(ex.SauceConfig)
				g.Assert(valid == nil).Equal(ex.ExpectedIsNil)
			})
		}
	})
}

// TestGetMetrics tests that the get metrics function works in saucelabs.go
func TestGetMetrics(t *testing.T) {
	g := goblin.Goblin(t)
	sc := fakeSauce()
	g.Describe("sauce getMetrics()", func() {
		res, err := getMetrics(fakeLog, fakeConfig, sc)
		g.It("should return metric data", func() {
			g.Assert(err).Equal(nil)
			g.Assert(len(res) > 0).Equal(true)
		})
		g.It("should have 'event_type' keys on everything", func() {
			for _, metric := range res {
				g.Assert(metric["event_type"] != nil).Equal(true)
			}
		})
		g.It("should have 'entity_name' keys on everything", func() {
			for _, metric := range res {
				g.Assert(metric["entity_name"] != nil).Equal(true)
			}
		})
		g.It("should have 'provider' keys on everything", func() {
			for _, metric := range res {
				g.Assert(metric["provider"] != nil).Equal(true)
			}
		})
	})
}

func TestGetUserList(t *testing.T) {
	g := goblin.Goblin(t)
	sc := fakeSauce()
	examples := []struct {
		Expected    []User
		CausesError bool
	}{
		{
			[]User{
				{
					"FIRSTUSER",
				},
				{
					"seconduser",
				},
			},
			false,
		},
	}
	g.Describe("GetUserList", func() {
		g.It("gets the user list", func() {
			for _, x := range examples {
				g.Describe("sauce GetUserList()", func() {
					res, err := sc.GetUserList()
					g.It("should return userlist data", func() {
						g.Assert(err != nil).Equal(x.CausesError)
						g.Assert(reflect.DeepEqual(x.Expected, res)).Equal(true)
					})
				})
			}
		})
	})
}

func TestGetUserActivity(t *testing.T) {
	g := goblin.Goblin(t)
	sc := fakeSauce()
	examples := []struct {
		Expected    Activity
		CausesError bool
	}{
		{
			Activity{
				SubAccounts: map[string]SubAccount{
					"steelers": SubAccount{7, 7, 0},
					"penguins": SubAccount{1, 1, 0},
					"pirates":  SubAccount{1, 2, 1},
				},
				Totals: SubAccount{9, 10, 1},
			},
			false,
		},
	}
	g.Describe("GetUserActivity", func() {
		g.It("gets the user list", func() {
			for _, x := range examples {
				g.Describe("sauce GetUserActivity()", func() {
					res, err := sc.GetUserActivity()
					g.It("should return user activity", func() {
						g.Assert(err != nil).Equal(x.CausesError)
						g.Assert(reflect.DeepEqual(x.Expected, res)).Equal(true)
					})
				})
			}
		})
	})
}

func TestGetConcurrency(t *testing.T) {
	g := goblin.Goblin(t)
	sc := fakeSauce()
	examples := []struct {
		Expected    Data
		CausesError bool
	}{
		{
			Data{
				Concurrency: map[string]TeamData{
					"self": TeamData{
						Current: Allocation{
							Overall: 4,
							Mac:     1,
							Manual:  0,
						},
						Remaining: Allocation{
							Overall: 0,
							Mac:     0,
							Manual:  0,
						},
					},
					"ancestor": TeamData{
						Current: Allocation{
							Overall: 4,
							Mac:     1,
							Manual:  0,
						},
						Remaining: Allocation{
							Overall: 0,
							Mac:     0,
							Manual:  0,
						},
					},
				},
			},
			false,
		},
	}

	g.Describe("GetConcurrency", func() {
		g.It("gets the user list", func() {
			for _, x := range examples {
				g.Describe("sauce GetConcurrency()", func() {
					res, err := sc.GetConcurrency()
					g.It("should return user concurrency", func() {
						g.Assert(err != nil).Equal(x.CausesError)
						g.Assert(reflect.DeepEqual(x.Expected, res)).Equal(true)
					})
				})
			}
		})
	})
}

func TestGetUsage(t *testing.T) {
	g := goblin.Goblin(t)
	sc := fakeSauce()
	examples := []struct {
		Expected    HistoryFormated
		CausesError bool
	}{
		{
			Expected: HistoryFormated{
				UserName: "testing",
				Usage: []UsageList{
					{
						Date: time.Date(2017, 7, 14, 0, 0, 0, 0, time.UTC),
						testInfoList: TestInfo{
							Executed: 24,
							Time:     6509,
						},
					},
					{
						Date: time.Date(2017, 7, 19, 0, 0, 0, 0, time.UTC),
						testInfoList: TestInfo{
							Executed: 2,
							Time:     266,
						},
					},
				},
			},
			CausesError: false,
		},
	}

	g.Describe("GetUsage", func() {
		g.It("gets the usage", func() {
			for _, x := range examples {
				g.Describe("sauce GetUsage()", func() {
					res, err := sc.GetUsage()
					g.It("should return user usage", func() {
						g.Assert(err != nil).Equal(x.CausesError)
						g.Assert(reflect.DeepEqual(x.Expected, res)).Equal(true)
					})
				})
			}
		})
	})
}

func TestGetErrors(t *testing.T) {
	g := goblin.Goblin(t)
	sc := fakeSauce()
	examples := []struct {
		Expected    Errors
		CausesError bool
	}{
		{
			Expected: Errors{
				Buckets: []BucketsList{
					{
						Name:  "Test did not see a new command for 90 seconds. Timing out.",
						Count: 1,
						Items: []ItemsList{
							{
								ID:           "k23j45bnkj236kj24klo34n67",
								Owner:        "johndoe",
								Name:         "Snickerdoodle",
								Build:        "0.5.0",
								CreationTime: "2017-10-23T06:28:37Z",
								StartTime:    "2017-10-23T06:28:37Z",
								EndTime:      "2017-10-23T07:05:37Z",
								Duration:     2220,
								Status:       "errored",
								Error:        "Test did not see a new command for 90 seconds. Timing out.",
								OS:           "OS X El Capitan (10.11)",
								Browser:      "Firefox 45.0",
								DetailsURL:   "https://saucelabs.com/rest/v1.1/johndoe/jobs/k23j45bnkj236kj24klo34n67",
							},
						},
					},
				},
			},
			CausesError: false,
		},
	}

	g.Describe("GetErrors", func() {
		g.It("gets the errors", func() {
			for _, x := range examples {
				g.Describe("sauce GetErrors()", func() {
					startDateString := "2017-10-22T12:00:00"
					endDateString := "2017-10-23T12:00:00"
					res, err := sc.GetErrors(startDateString, endDateString)
					g.It("should return user errors", func() {
						g.Assert(err != nil).Equal(x.CausesError)
						g.Assert(reflect.DeepEqual(x.Expected, res)).Equal(true)
					})
				})
			}
		})
	})
}

func fakeSauce() *SauceClient {
	sc, scErr := NewSauceClient(fakeConfig)
	if scErr != nil {
		panic(scErr)
	}

	fakeSauceTransport := httpmock.NewMockTransport()
	registerResponders(fakeSauceTransport)
	sc.Client = &http.Client{
		Transport: fakeSauceTransport,
	}
	return sc
}

// hash map of HTTP requests to mock
func registerResponders(transport *httpmock.MockTransport) {
	responses := []struct {
		Method   string
		Endpoint string
		Code     int
		Response string
	}{
		{
			"GET",
			"https://saucelabs.com/rest/v1/users/test-user/activity",
			200,
			`{"subaccounts":{"steelers":{"in progress":7,"all":7,"queued":0},"penguins":{"in progress":1,"all":1,"queued":0},"pirates":{"in progress":1,"all":2,"queued":1}},"totals":{"in progress":9,"all":10,"queued":1}}`,
		},
		{
			"GET",
			"https://saucelabs.com/rest/v1/users/test-user/concurrency",
			200,
			`{"timestamp":8.7712928E8,"concurrency":{"self":{"username":"testing","current":{"manual":0,"mac":1,"overall":4},"allowed":{"manual":43,"mac":43,"overall":43}},"ancestor":{"username":"testing","current":{"manual":0,"mac":1,"overall":4},"allowed":{"manual":43,"mac":43,"overall":43}}}}`,
		},
		{
			"GET",
			"https://saucelabs.com/rest/v1/users/test-user/subaccounts",
			200,
			`[{"username":"FIRSTUSER","vm_lockdown":false,"new_email":null,"last_name":"Doe","parent":"gd_automation","subaccount_limit":null,"children_count":0,"creation_time":1465484232,"user_type":"subaccount","monthly_minutes":{"manual":"infinite","automated":999999},"prevent_emails":[],"is_admin":null,"manual_minutes":"infinite","can_run_manual":true,"concurrency_limit":{"mac":70,"scout":3,"overall":3,"real_device":0},"is_public":false,"id":"FIRSTUSER","access_key":"12345","first_name":"John","require_full_name":true,"verified":false,"name":"John Doe","subscribed":true,"title":null,"terminating_subscription":false,"is_sso":false,"entity_type":null,"tunnel_concurrency_limit":70,"allow_integrations_page":true,"last_login":null,"ancestor_concurrency_limit":{"mac":70,"scout":70,"overall":70,"real_device":0},"ancestor_allows_subaccounts":false,"domain":null,"ancestor":"gd_automation","minutes":980420,"email":"johndoe@gannett.com"},{"username":"seconduser","vm_lockdown":false,"new_email":null,"last_name":null,"parent":"gd_automation","subaccount_limit":null,"last_test":1502724106,"children_count":0,"creation_time":1502719650,"user_type":"subaccount","monthly_minutes":{"manual":"infinite","automated":999999},"prevent_emails":[],"is_admin":null,"manual_minutes":"infinite","can_run_manual":true,"concurrency_limit":{"mac":70,"scout":3,"overall":3,"real_device":0},"is_public":false,"id":"seconduser","access_key":"123743","first_name":null,"require_full_name":true,"verified":true,"name":"Jane Doe","subscribed":true,"title":null,"terminating_subscription":false,"is_sso":false,"entity_type":null,"tunnel_concurrency_limit":70,"allow_integrations_page":true,"last_login":1502722851,"ancestor_concurrency_limit":{"mac":70,"scout":70,"overall":70,"real_device":0},"ancestor_allows_subaccounts":false,"domain":null,"ancestor":"gd_automation","minutes":980420,"email":"janedoe@gannett.com"}]`,
		},
		{
			"GET",
			"https://saucelabs.com/rest/v1/users/test-user/usage",
			200,
			`{"usage":[["2017-7-14",[24,6509]],["2017-7-19",[2,266]]],"username":"testing"}`,
		},
		{
			"GET",
			"https://saucelabs.com/rest/v1/analytics/trends/errors?end=2017-10-23T12%3A00%3A00Z&scope=organization&start=2017-10-22T12%3A00%3A00Z",
			200,
			`{"meta":{"status":200},"buckets":[{"name":"Test did not see a new command for 90 seconds. Timing out.","count":1,"items":[{"id":"k23j45bnkj236kj24klo34n67","owner":"johndoe","ancestor":"testing","name":"Snickerdoodle","build":"0.5.0","creation_time":"2017-10-23T06:28:37Z","start_time":"2017-10-23T06:28:37Z","end_time":"2017-10-23T07:05:37Z","duration":2220,"status":"errored","error":"Test did not see a new command for 90 seconds. Timing out.","os":"OS X El Capitan (10.11)","browser":"Firefox 45.0","details_url":"https://saucelabs.com/rest/v1.1/johndoe/jobs/k23j45bnkj236kj24klo34n67"}],"has_more":false}],"all_items_count":1}`,
		},
	}

	for r := range responses {
		match := responses[r]
		transport.RegisterResponder(match.Method, match.Endpoint, func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(match.Code, match.Response)
			resp.Header.Add("Content-Type", "application/json")
			resp.Header.Add("X-Sauce", "mock")
			if testing.Verbose() {
				fmt.Println("httpmock: match", req.Method, req.URL, "->", match.Endpoint)
			}
			return resp, nil
		})
	}

	transport.RegisterNoResponder(func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(501, "{}")
		resp.Header.Add("X-Sauce", "mock")
		if testing.Verbose() {
			fmt.Println("httpmock: no match", req.Method, req.URL)
		}
		return resp, nil
	})
}
