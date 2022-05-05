package yop_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/hasanozgan/yaml-object-parser"
	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	yop.AddObjectName(
		"or",
		"and",
		"user",
		"service",
		"location",
		"opening-hours",
		"relationship",
	)

	tests := map[string]struct {
		yamlSection    []byte
		expectedObject *yop.Object
		expectedError  error
	}{
		"rule string format": {
			yamlSection:    yamlStringObject(),
			expectedObject: expectedStringObject(),
		},
		"rule object format": {
			yamlSection:    yamlObject(),
			expectedObject: expectedObject(),
		},
		"rule list format": {
			yamlSection:    yamlListObject(),
			expectedObject: expectedListObject(),
		},
		"rule nested format": {
			yamlSection:    yamlNestedObject(),
			expectedObject: expectedNestedObject(),
		},
		"rule nested format with object": {
			yamlSection:    yamlNestedObjectWithObject(),
			expectedObject: expectedNestedObjectWithObject(),
		},
		"error rule max depth limit exceeded": {
			yamlSection:   yamlErrorObjectMaxDepthLimitExceeded(),
			expectedError: yop.ErrMaxDepthLimitExceeded,
		},
		"error null rule": {
			yamlSection:   yamlErrorNullObject(),
			expectedError: yop.ErrObjectUndefined,
		},
		"error empty rule": {
			yamlSection:   yamlErrorEmptyObjectName(),
			expectedError: yop.ErrObjectUndefined,
		},
		"error rule not found": {
			yamlSection:   yamlErrorNotFoundObject(),
			expectedError: yop.ErrObjectNotFound,
		},
		"error rule list not acceptable": {
			yamlSection:   yamlErrorListNotAcceptable(),
			expectedError: yop.ErrObjectNotAcceptable,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := &testYamlSection{}
			actualError := yaml.Unmarshal(test.yamlSection, actual)

			assertError(t, test.expectedError, actualError)
			require.Equal(t, test.expectedObject, actual.Rule)
		})
	}
}

func yamlStringObject() []byte {
	return []byte("rule: user")
}

func expectedStringObject() *yop.Object {
	return &yop.Object{
		Name: "user",
	}
}

func yamlObject() []byte {
	return []byte(strings.TrimSpace(`
rule:
  location:
    region: uk
`))
}

func expectedObject() *yop.Object {
	return &yop.Object{
		Name:      "location",
		Arguments: `{"region":"uk"}`,
	}
}

func yamlListObject() []byte {
	return []byte(strings.TrimSpace(`
rule:
  and:
    - user
    - service
`))
}

func expectedListObject() *yop.Object {
	return &yop.Object{
		Name: "and",
		Children: []*yop.Object{
			{Name: "user", Level: 1},
			{Name: "service", Level: 1},
		},
	}
}

func yamlNestedObject() []byte {
	return []byte(strings.TrimSpace(`
rule:
  or:
    - user
    - and:
        - service
        - opening-hours
`))
}

func expectedNestedObject() *yop.Object {
	return &yop.Object{
		Name: "or",
		Children: []*yop.Object{
			{Name: "user", Level: 1},
			{
				Name:  "and",
				Level: 1,
				Children: []*yop.Object{
					{Name: "service", Level: 2},
					{Name: "opening-hours", Level: 2},
				},
			},
		},
	}
}

func yamlNestedObjectWithObject() []byte {
	return []byte(strings.TrimSpace(`
rule:
  or:
    - user
    - and:
        - service
        - opening-hours:
            opening: 10:00
            closing: 20:00
        - relationship:
            levels:
              - comprehensive
              - parent
`))
}

func expectedNestedObjectWithObject() *yop.Object {
	return &yop.Object{
		Name: "or",
		Children: []*yop.Object{
			{Name: "user", Level: 1},
			{
				Name:  "and",
				Level: 1,
				Children: []*yop.Object{
					{Name: "service", Level: 2},
					{
						Name:      "opening-hours",
						Level:     2,
						Arguments: `{"closing":"20:00","opening":"10:00"}`,
					},
					{
						Name:      "relationship",
						Level:     2,
						Arguments: `{"levels":["comprehensive","parent"]}`,
					},
				},
			},
		},
	}
}

func yamlErrorNullObject() []byte {
	return []byte("rule: ~")
}

func yamlErrorEmptyObjectName() []byte {
	return []byte(`rule: ""`)
}

func yamlErrorNotFoundObject() []byte {
	return []byte(`rule: "invalid"`)
}

func yamlErrorObjectMaxDepthLimitExceeded() []byte {
	return []byte(strings.TrimSpace(`
rule:
  and: # level 0
    - or: # level 1
      - and: # level 2
        - or: # level 3 -> max depth limit exceeded
`))
}

func yamlErrorListNotAcceptable() []byte {
	return []byte(strings.TrimSpace(`
rule:
  - user
  - location:
      region: uk
`))
}

func assertError(t *testing.T, expected, actual error) {
	if expected == nil {
		require.NoError(t, actual)
	} else {
		require.Contains(t, actual.Error(), expected.Error())
	}
}

type testYamlSection struct {
	Rule *yop.Object `json:"rule"`
}

func (t *testYamlSection) UnmarshalJSON(b []byte) (err error) {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(b, &result); err != nil {
		return err
	}

	for key, value := range result {
		if key == "rule" {
			if t.Rule, err = yop.ParseObjectFromJSON(value); err != nil {
				return err
			}
		}
	}
	return nil
}
