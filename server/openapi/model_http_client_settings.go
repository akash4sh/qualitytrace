/*
 * TraceTest
 *
 * OpenAPI definition for TraceTest endpoint and resources
 *
 * API version: 0.2.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type HttpClientSettings struct {
	Url string `json:"url,omitempty"`

	Headers map[string]string `json:"headers,omitempty"`

	Tls Tls `json:"tls,omitempty"`

	Auth HttpAuth `json:"auth,omitempty"`
}

// AssertHttpClientSettingsRequired checks if the required fields are not zero-ed
func AssertHttpClientSettingsRequired(obj HttpClientSettings) error {
	if err := AssertTlsRequired(obj.Tls); err != nil {
		return err
	}
	if err := AssertHttpAuthRequired(obj.Auth); err != nil {
		return err
	}
	return nil
}

// AssertRecurseHttpClientSettingsRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of HttpClientSettings (e.g. [][]HttpClientSettings), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseHttpClientSettingsRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aHttpClientSettings, ok := obj.(HttpClientSettings)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertHttpClientSettingsRequired(aHttpClientSettings)
	})
}