package azure

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/trufflesecurity/trufflehog/v3/pkg/detectors"
	"github.com/trufflesecurity/trufflehog/v3/pkg/engine/ahocorasick"
)

var (
	validPattern = `
	azure credentials:
		azureClientID: clientid9304d5df4-aac1-6117-552c-7f70c89a40d9
		azureTenant: tenant_idid9304d5df4-aac1-6117-552c-7f70c89a40d9
		azureClientSecret: clientsecretY_0w|[cGpan41k6ng.ol414sp4ccw2v_rkfmbs537i
	`
	invalidPattern = `
	azure credentials:
		azureClientID: 9304d5df4-aac1-6117-552c-7f70c89a
		azureTenant: id9304d5df4-aac1-6117-55-7f70c89a40d9
		azureClientSecret: Y_0w|[cGpan41k6ng.
	`
)

func TestAzure_Pattern(t *testing.T) {
	d := Scanner{}
	ahoCorasickCore := ahocorasick.NewAhoCorasickCore([]detectors.Detector{d})

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "valid pattern",
			input: validPattern,
			want:  []string{"304d5df4-aac1-6117-552c-7f70c89a40d9cGpan41k6ng.ol414sp4ccw2v_rkfmbs53304d5df4-aac1-6117-552c-7f70c89a40d9"},
		},
		{
			name:  "invalid pattern",
			input: invalidPattern,
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matchedDetectors := ahoCorasickCore.FindDetectorMatches([]byte(test.input))
			if len(matchedDetectors) == 0 {
				t.Errorf("keywords '%v' not matched by: %s", d.Keywords(), test.input)
				return
			}

			results, err := d.FromData(context.Background(), false, []byte(test.input))
			if err != nil {
				t.Errorf("error = %v", err)
				return
			}

			if len(results) != len(test.want) {
				if len(results) == 0 {
					t.Errorf("did not receive result")
				} else {
					t.Errorf("expected %d results, only received %d", len(test.want), len(results))
				}
				return
			}

			actual := make(map[string]struct{}, len(results))
			for _, r := range results {
				if len(r.RawV2) > 0 {
					actual[string(r.RawV2)] = struct{}{}
				} else {
					actual[string(r.Raw)] = struct{}{}
				}
			}
			expected := make(map[string]struct{}, len(test.want))
			for _, v := range test.want {
				expected[v] = struct{}{}
			}

			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s diff: (-want +got)\n%s", test.name, diff)
			}
		})
	}
}