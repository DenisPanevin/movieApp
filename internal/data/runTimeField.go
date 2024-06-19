package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errInvalidRuntimeFormat =errors.New("invalid runtime format")

type RunTime int32

func (r RunTime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedValue := strconv.Quote(jsonValue)
	return []byte(quotedValue), nil
}

func (r *RunTime)UnmarshalJSON(jsonValue []byte)error{

	unqouted,err:=strconv.Unquote(string(jsonValue))
	if err != nil {
		return errInvalidRuntimeFormat
	}
	parts:=strings.Split(unqouted," ")
	if len(parts) != 2 || parts[1] != "mins"{
		return errInvalidRuntimeFormat
	}
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return errInvalidRuntimeFormat
	}
	*r=RunTime(i)

	return nil

}
