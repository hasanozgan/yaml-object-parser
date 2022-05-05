# Yaml Object Parser

### Object Format

```go
type Object struct {
	Name      string
	Level     int
	Children  []*Object
	Arguments string
}
```

### Example Struct

```go
// set max depth limit
yop.SetMaxDepthLimit(2)

type RuleSection struct {
	Rule *yop.Object `json:"rule"`
}

func (t *RuleSection) UnmarshalJSON(b []byte) (err error) {
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
```

### Example Yaml File

```yaml

error-null-object:
  rule: ~

error-empty-object:
  rule: ""

error-not-found:
  rule: invalid

error-tree-level-limit-exceeded:
    rule:
      and:
        - or:
          - and:
            - or:
              - and:
                - or:
                  - and:
                    - or:
                      - and:
                        - self
                        - location:
                            abc: 123

error-object-list-not-acceptable:
  rule:
    - self
    - location:
        region: uk

string-object:
  rule: test

object-object:
  rule:
    location:
      region: uk

list-object:
  rule:
    and:
      - self
      - location:
          region: us

nested-object:
  rule:
    or:
      - self
      - and:
          - location
          - time

nested-object-with-object:
  rule:
    or:
      - self
      - and:
          - location
          - opening-hours:
              opening: 10:00
              closing: 20:00
          - relationship:
              levels:
                - comprehensive
                - parent


```