/*
Copyright 2025-2026 Trident Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to
// test unexported functions.
package datatypes

import (
	"strings"
	"testing"
)

// TestNewFacetExplicitTimezone verifies that NewFacetExplicitTimezone correctly
// initializes the facet's name, fixed flag, and value.
func TestNewFacetExplicitTimezone(t *testing.T) {
	tests := []struct {
		name     string
		val      ExplicitTimezoneValue
		fixed    bool
		wantName FacetName
	}{
		{
			name:     "required and fixed",
			val:      TimezoneRequired,
			fixed:    true,
			wantName: NameExplicitTimezone,
		},
		{
			name:     "prohibited and not fixed",
			val:      TimezoneProhibited,
			fixed:    false,
			wantName: NameExplicitTimezone,
		},
		{
			name:     "optional and fixed",
			val:      TimezoneOptional,
			fixed:    true,
			wantName: NameExplicitTimezone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facet := NewFacetExplicitTimezone(tt.val, tt.fixed)

			if facet.Name != tt.wantName {
				t.Errorf("NewFacetExplicitTimezone().Name = %v, want %v", facet.Name, tt.wantName)
			}
			if facet.Fixed != tt.fixed {
				t.Errorf("NewFacetExplicitTimezone().Fixed = %v, want %v", facet.Fixed, tt.fixed)
			}
			if facet.Value != tt.val {
				t.Errorf("NewFacetExplicitTimezone().Value = %v, want %v", facet.Value, tt.val)
			}
		})
	}
}

// TestFacetExplicitTimezoneCheck verifies the Check method of explicit timezone
// facets under various conditions (required, prohibited, optional, and unsupported types).
func TestFacetExplicitTimezoneCheck(t *testing.T) {
	zonedDate, err := NewDate(2025, 1, 1, GetUTC())
	if err != nil {
		t.Fatalf("unexpected error setting up zoned date: %v", err)
	}

	unzonedDate, err := NewDate(2025, 1, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error setting up unzoned date: %v", err)
	}

	nonTimezoneVal := String("test string")

	tests := []struct {
		name           string
		facetValue     ExplicitTimezoneValue
		inputVal       XSDValue
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:           "unsupported datatype",
			facetValue:     TimezoneRequired,
			inputVal:       nonTimezoneVal,
			wantErr:        true,
			expectedErrMsg: "does not support explicitTimezone facet",
		},
		{
			name:           "required: absent timezone error",
			facetValue:     TimezoneRequired,
			inputVal:       unzonedDate,
			wantErr:        true,
			expectedErrMsg: "explicitTimezone violation: timezone is required but was absent",
		},
		{
			name:       "required: present timezone success",
			facetValue: TimezoneRequired,
			inputVal:   zonedDate,
			wantErr:    false,
		},
		{
			name:           "prohibited: present timezone error",
			facetValue:     TimezoneProhibited,
			inputVal:       zonedDate,
			wantErr:        true,
			expectedErrMsg: "explicitTimezone violation: timezone is prohibited but was present",
		},
		{
			name:       "prohibited: absent timezone success",
			facetValue: TimezoneProhibited,
			inputVal:   unzonedDate,
			wantErr:    false,
		},
		{
			name:       "optional: present timezone success",
			facetValue: TimezoneOptional,
			inputVal:   zonedDate,
			wantErr:    false,
		},
		{
			name:       "optional: absent timezone success",
			facetValue: TimezoneOptional,
			inputVal:   unzonedDate,
			wantErr:    false,
		},
		{
			name:       "unrecognized timezone setting passthrough",
			facetValue: ExplicitTimezoneValue("unknown"),
			inputVal:   zonedDate,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facet := NewFacetExplicitTimezone(tt.facetValue, false)
			checkErr := facet.Check(tt.inputVal)

			if (checkErr != nil) != tt.wantErr {
				t.Fatalf("FacetExplicitTimezone.Check() error = %v, wantErr %v", checkErr, tt.wantErr)
			}
			if tt.wantErr && tt.expectedErrMsg != "" {
				if !strings.Contains(checkErr.Error(), tt.expectedErrMsg) {
					t.Errorf(
						"FacetExplicitTimezone.Check() error message = %q, expected substring %q",
						checkErr.Error(),
						tt.expectedErrMsg,
					)
				}
			}
		})
	}
}
