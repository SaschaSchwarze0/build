// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

// OverrideParams allows to override an existing list of parameters with a second list,
// as long as their entry names matches
func OverrideParams(originalParams []buildv1alpha1.Param, overrideParams []buildv1alpha1.Param) []buildv1alpha1.Param {
	if len(overrideParams) == 0 {
		return originalParams
	}

	if len(originalParams) == 0 && len(overrideParams) > 0 {
		return overrideParams
	}

	// override params that matches by name in both originalParams and overrideParams
	for i, o := range originalParams {
		for _, p := range overrideParams {
			if o.Name == p.Name {
				originalParams[i] = buildv1alpha1.Param{
					Name:  p.Name,
					Value: p.Value,
				}
			}
		}
	}

	// with a modified originalParams list, add parameters that only
	// the overrideParams list contains
	for _, p := range overrideParams {
		auxFlag := true
		for _, original := range originalParams {
			if p.Name == original.Name {
				auxFlag = false
			}
		}
		if auxFlag {
			originalParams = append(originalParams, p)
		}
	}

	return originalParams
}
