package scripts

import (
	"fmt"
)

type DriverCellScriptResult struct {
	Status  int64
	OldCell string
}

type RemoveDriverCellScriptResult struct {
	Status int64
	Cell   string
}

func DecodeDriverCellScriptResult(
	raw interface{},
) (DriverCellScriptResult, error) {

	values, ok := raw.([]interface{})
	if !ok {
		return DriverCellScriptResult{},
			fmt.Errorf("invalid driver cell script response type")
	}

	if len(values) != 2 {
		return DriverCellScriptResult{},
			fmt.Errorf(
				"invalid driver cell script response length: %d",
				len(values),
			)
	}

	status, err := asInt64(values[0])
	if err != nil {
		return DriverCellScriptResult{}, err
	}

	oldCell, err := asString(values[1])
	if err != nil {
		return DriverCellScriptResult{}, err
	}

	switch status {
	case DriverCellScriptUnchanged,
		DriverCellScriptMoved,
		DriverCellScriptAdded:

	default:
		return DriverCellScriptResult{},
			fmt.Errorf(
				"unknown driver cell script status: %d",
				status,
			)
	}

	return DriverCellScriptResult{
		Status:  status,
		OldCell: oldCell,
	}, nil
}

func DecodeRemoveDriverCellScriptResult(
	raw interface{},
) (RemoveDriverCellScriptResult, error) {

	values, ok := raw.([]interface{})
	if !ok {
		return RemoveDriverCellScriptResult{},
			fmt.Errorf("invalid remove driver cell script response type")
	}

	if len(values) != 2 {
		return RemoveDriverCellScriptResult{},
			fmt.Errorf(
				"invalid remove driver cell script response length: %d",
				len(values),
			)
	}

	status, err := asInt64(values[0])
	if err != nil {
		return RemoveDriverCellScriptResult{}, err
	}

	cell, err := asString(values[1])
	if err != nil {
		return RemoveDriverCellScriptResult{}, err
	}

	switch status {
	case DriverCellRemoveNotFound,
		DriverCellRemoved:

	default:
		return RemoveDriverCellScriptResult{},
			fmt.Errorf(
				"unknown remove driver cell status: %d",
				status,
			)
	}

	return RemoveDriverCellScriptResult{
		Status: status,
		Cell:   cell,
	}, nil
}

func asInt64(v interface{}) (int64, error) {
	value, ok := v.(int64)
	if !ok {
		return 0,
			fmt.Errorf("expected int64 but got %T", v)
	}

	return value, nil
}

func asString(v interface{}) (string, error) {
	value, ok := v.(string)
	if !ok {
		return "",
			fmt.Errorf("expected string but got %T", v)
	}

	return value, nil
}
