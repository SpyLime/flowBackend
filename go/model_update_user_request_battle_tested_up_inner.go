// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Flow Learning - OpenAPI 3.1
 *
 * api for flow learning
 *
 * API version: 1.0.0
 * Contact: floTeam@gmail.com
 */

package openapi


import (
	"time"
)



type UpdateUserRequestBattleTestedUpInner struct {

	Topic string `json:"topic,omitempty"`

	Title string `json:"title,omitempty"`

	NodeId time.Time `json:"nodeId,omitempty"`
}

// AssertUpdateUserRequestBattleTestedUpInnerRequired checks if the required fields are not zero-ed
func AssertUpdateUserRequestBattleTestedUpInnerRequired(obj UpdateUserRequestBattleTestedUpInner) error {
	return nil
}

// AssertUpdateUserRequestBattleTestedUpInnerConstraints checks if the values respects the defined constraints
func AssertUpdateUserRequestBattleTestedUpInnerConstraints(obj UpdateUserRequestBattleTestedUpInner) error {
	return nil
}
