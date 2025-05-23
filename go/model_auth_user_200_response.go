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




type AuthUser200Response struct {

	IsAuth bool `json:"isAuth"`

	Role int32 `json:"role"`
}

// AssertAuthUser200ResponseRequired checks if the required fields are not zero-ed
func AssertAuthUser200ResponseRequired(obj AuthUser200Response) error {
	elements := map[string]interface{}{
		"isAuth": obj.IsAuth,
		"role": obj.Role,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertAuthUser200ResponseConstraints checks if the values respects the defined constraints
func AssertAuthUser200ResponseConstraints(obj AuthUser200Response) error {
	return nil
}
