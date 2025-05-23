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




type AddTopic200Response struct {

	Topic GetTopics200ResponseInner `json:"topic"`

	NodeData AddTopic200ResponseNodeData `json:"nodeData"`
}

// AssertAddTopic200ResponseRequired checks if the required fields are not zero-ed
func AssertAddTopic200ResponseRequired(obj AddTopic200Response) error {
	elements := map[string]interface{}{
		"topic": obj.Topic,
		"nodeData": obj.NodeData,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	if err := AssertGetTopics200ResponseInnerRequired(obj.Topic); err != nil {
		return err
	}
	if err := AssertAddTopic200ResponseNodeDataRequired(obj.NodeData); err != nil {
		return err
	}
	return nil
}

// AssertAddTopic200ResponseConstraints checks if the values respects the defined constraints
func AssertAddTopic200ResponseConstraints(obj AddTopic200Response) error {
	if err := AssertGetTopics200ResponseInnerConstraints(obj.Topic); err != nil {
		return err
	}
	if err := AssertAddTopic200ResponseNodeDataConstraints(obj.NodeData); err != nil {
		return err
	}
	return nil
}
