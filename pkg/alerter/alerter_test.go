package alerter

import (
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resource2 "github.com/cloudskiff/driftctl/test/resource"
)

type fakeAlert struct {
	message              string
	shouldIgnoreResource bool
}

func newFakeAlert(message string, shouldIgnoreResource bool) fakeAlert {
	return fakeAlert{message, shouldIgnoreResource}
}

func (f fakeAlert) Message() string {
	return f.message
}

func (f fakeAlert) ShouldIgnoreResource() bool {
	return f.shouldIgnoreResource
}

func TestAlerter_Alert(t *testing.T) {
	cases := []struct {
		name     string
		alerts   Alerts
		expected Alerts
	}{
		{
			name:     "TestNoAlerts",
			alerts:   nil,
			expected: Alerts{},
		},
		{
			name: "TestWithSingleAlert",
			alerts: Alerts{
				"fakeres.foobar": []Alert{
					newFakeAlert("This is an alert", false),
				},
			},
			expected: Alerts{
				"fakeres.foobar": []Alert{
					newFakeAlert("This is an alert", false),
				},
			},
		},
		{
			name: "TestWithMultipleAlerts",
			alerts: Alerts{
				"fakeres.foobar": []Alert{
					newFakeAlert("This is an alert", false),
					newFakeAlert("This is a second alert", true),
				},
				"fakeres.barfoo": []Alert{
					newFakeAlert("This is a third alert", true),
				},
			},
			expected: Alerts{
				"fakeres.foobar": []Alert{
					newFakeAlert("This is an alert", false),
					newFakeAlert("This is a second alert", true),
				},
				"fakeres.barfoo": []Alert{
					newFakeAlert("This is a third alert", true),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			alerter := NewAlerter()

			for k, v := range c.alerts {
				for _, a := range v {
					alerter.SendAlert(k, a)
				}
			}

			if eq := reflect.DeepEqual(alerter.Retrieve(), c.expected); !eq {
				t.Errorf("Got %+v, expected %+v", alerter.Retrieve(), c.expected)
			}
		})
	}
}

func TestAlerter_IgnoreResources(t *testing.T) {
	cases := []struct {
		name     string
		alerts   Alerts
		resource resource.Resource
		expected bool
	}{
		{
			name:   "TestNoAlerts",
			alerts: Alerts{},
			resource: &resource2.FakeResource{
				Type: "fakeres",
				Id:   "foobar",
			},
			expected: false,
		},
		{
			name: "TestShouldNotBeIgnoredWithAlerts",
			alerts: Alerts{
				"fakeres": {
					newFakeAlert("Should not be ignored", false),
				},
				"fakeres.foobar": {
					newFakeAlert("Should not be ignored", false),
				},
				"fakeres.barfoo": {
					newFakeAlert("Should not be ignored", false),
				},
				"other.resource": {
					newFakeAlert("Should not be ignored", false),
				},
			},
			resource: &resource2.FakeResource{
				Type: "fakeres",
				Id:   "foobar",
			},
			expected: false,
		},
		{
			name: "TestShouldBeIgnoredWithAlertsOnWildcard",
			alerts: Alerts{
				"fakeres": {
					newFakeAlert("Should be ignored", true),
				},
				"other.foobaz": {
					newFakeAlert("Should be ignored", true),
				},
				"other.resource": {
					newFakeAlert("Should not be ignored", false),
				},
			},
			resource: &resource2.FakeResource{
				Type: "fakeres",
				Id:   "foobar",
			},
			expected: true,
		},
		{
			name: "TestShouldBeIgnoredWithAlertsOnResource",
			alerts: Alerts{
				"fakeres": {
					newFakeAlert("Should be ignored", true),
				},
				"other.foobaz": {
					newFakeAlert("Should be ignored", true),
				},
				"other.resource": {
					newFakeAlert("Should not be ignored", false),
				},
			},
			resource: &resource2.FakeResource{
				Type: "other",
				Id:   "foobaz",
			},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			alerter := NewAlerter()
			alerter.SetAlerts(c.alerts)
			if got := alerter.IsResourceIgnored(c.resource); got != c.expected {
				t.Errorf("Got %+v, expected %+v", got, c.expected)
			}
		})
	}
}
