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




type Topic struct {

	Title string `json:"title"`
}

// AssertTopicRequired checks if the required fields are not zero-ed
func AssertTopicRequired(obj Topic) error {
	elements := map[string]interface{}{
		"title": obj.Title,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertTopicConstraints checks if the values respects the defined constraints
func AssertTopicConstraints(obj Topic) error {
	return nil
}
