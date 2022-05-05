package yop

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var maxDepthLimit = 2

var (
	errObject                = errors.New("parameter")
	ErrObjectNotFound        = fmt.Errorf("%w not found", errObject)
	ErrObjectNotAcceptable   = fmt.Errorf("%w not acceptable", errObject)
	ErrObjectUndefined       = fmt.Errorf("%w undefined", errObject)
	ErrMaxDepthLimitExceeded = fmt.Errorf("%w max depth limit (%d) exceeded", errObject, maxDepthLimit)
)

type Object struct {
	Name      string
	Level     int
	Children  []*Object
	Arguments string
}

var validObjectNames = []string{}

func SetMaxDepthLimit(limit int) {
	maxDepthLimit = limit
}

func AddObjectName(names ...string) {
	if len(names) > 0 {
		validObjectNames = append(validObjectNames, names...)
	}
}

func RemoveObjectName(name string) {
	result := []string{}
	for _, objectName := range validObjectNames {
		if !strings.EqualFold(objectName, name) {
			result = append(result, objectName)
		}
	}
	validObjectNames = result
}

func ParseObjectFromJSON(jsonObjectSection []byte) (*Object, error) {
	if err := isObjectValid(jsonObjectSection); err != nil {
		return nil, err
	}

	obj := &Object{}
	if err := parseObject(obj, jsonObjectSection); err != nil {
		return nil, err
	}
	return obj, nil
}

func isObjectValid(b []byte) error {
	if string(b) == "\"\"" {
		return fmt.Errorf("%w: object has empty string", ErrObjectUndefined)
	}
	if string(b) == "null" {
		return fmt.Errorf("%w: object is null object", ErrObjectUndefined)
	}

	if string(b[0]) == "[" {
		return fmt.Errorf("%w: %s, please choose one", ErrObjectNotAcceptable, b)
	}

	return nil
}

func parseObject(parent *Object, b []byte) (err error) {
	if parent.Level > maxDepthLimit {
		return ErrMaxDepthLimitExceeded
	}

	err = parseObjectFromString(parent, b)
	if err == nil || errors.Is(err, errObject) {
		return err
	}

	err = parseObjectWithChildren(parent, b)
	if err == nil || errors.Is(err, errObject) {
		return err
	}

	return parseObjectWithArguments(parent, b)
}

func parseObjectFromString(parent *Object, b []byte) error {
	var objectName string
	if err := json.Unmarshal(b, &objectName); err != nil {
		return err // ignore error, json: cannot unmarshal object into Go value of type
	}

	if err := isSupported(objectName); err != nil {
		return err
	}
	parent.Name = objectName
	return nil
}

func parseObjectWithChildren(parent *Object, b []byte) error {
	var listObject map[string][]json.RawMessage
	if err := json.Unmarshal(b, &listObject); err != nil {
		return err // ignore error, json: cannot unmarshal object into Go value of type
	}
	for objectName, children := range listObject {
		if err := isSupported(objectName); err != nil {
			return err
		}

		parent.Name = objectName
		for _, child := range children {
			childObject := &Object{
				Level: parent.Level + 1,
			}
			if err := parseObject(childObject, child); err != nil {
				return err
			}
			parent.Children = append(parent.Children, childObject)
		}
	}

	return nil
}

func parseObjectWithArguments(parent *Object, b []byte) error {
	var objectObject map[string]json.RawMessage
	if err := json.Unmarshal(b, &objectObject); err != nil {
		return err // ignore error, json: cannot unmarshal object into Go value of type
	}

	for objectName, arguments := range objectObject {
		if err := isSupported(objectName); err != nil {
			return err
		}
		parent.Name = objectName
		parent.Arguments = string(arguments)
	}
	return nil
}

func isSupported(key string) error {
	for _, name := range validObjectNames {
		if key == name {
			return nil
		}
	}
	return fmt.Errorf("'%s' %w", key, ErrObjectNotFound)
}
