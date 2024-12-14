package core

import (
	"errors"
	"fmt"
)

type Resp struct {
	data []byte
	pos  int
}

func GetNewResp(data []byte) Resp {
	return Resp{data: data, pos: 0}
}

func (resp *Resp) DecodeArrayString() ([]string, error) {
	res, err := resp.Decode()
	if err != nil {
		return nil, err
	}
	tmp := res.([]interface{})
	tokens := make([]string, len(tmp))
	for i := range tokens {
		tokens[i] = tmp[i].(string)
	}
	return tokens, nil
}

func (resp *Resp) Decode() (interface{}, error) {
	if len(resp.data) == 0 {
		return nil, errors.New("no data")
	}
	value, err := resp.decodeOne()
	return value, err
}

func (resp *Resp) decodeOne() (interface{}, error) {
	if len(resp.data) == 0 {
		return nil, errors.New("no data")
	}

	startPos := resp.pos
	resp.pos++

	switch resp.data[startPos] {
	case '+':
		return resp.readSimpleString()
	case '-':
		return resp.readError()
	case ':':
		return resp.readInt64()
	case '$':
		return resp.readBulkString()
	case '*':
		return resp.readArray()
	}

	return nil, errors.New("invalid start of command")
}

func (resp *Resp) readInt64() (int64, error) {
	var value int64 = 0

	for ; resp.pos < len(resp.data); resp.pos++ {
		if resp.data[resp.pos] == '\r' {
			resp.pos += 2
			return value, nil
		}
		value = (value * 10) + int64(resp.data[resp.pos]-'0')
	}
	return 0, errors.New("invalid int64")
}

func (resp *Resp) readSimpleString() (string, error) {
	var startPos, endPos int = resp.pos, 0
	for ; resp.pos < len(resp.data); resp.pos++ {
		if resp.data[resp.pos] == '\r' {
			endPos = resp.pos
			break
		}
	}
	str := string(resp.data[startPos:endPos])
	resp.pos += 2
	return str, nil
}

func (resp *Resp) readError() (string, error) {
	return resp.readSimpleString()
}

func (resp *Resp) readBulkString() (string, error) {
	length, err := resp.readInt64()
	if err != nil {
		return "", err
	}
	intLength := int(length)
	str := string(resp.data[resp.pos : resp.pos+intLength])
	resp.pos += intLength + 2
	return str, nil
}

func (resp *Resp) readArray() ([]interface{}, error) {
	elementCount, err := resp.readInt64()
	if err != nil {
		return nil, err
	}

	elems := make([]interface{}, elementCount)

	for i := range elems {
		elem, err := resp.decodeOne()
		if err != nil {
			return nil, err
		}
		elems[i] = elem
	}

	return elems, nil
}

func Encode(data interface{}, isSimpleString bool) ([]byte, error) {
	// d is created since len doesn't accept interface{}. Go handles the type of d because of the switch statement.
	switch d := data.(type) {
	case string:
		if isSimpleString {
			return []byte(fmt.Sprintf("+%s\r\n", d)), nil
		} else {
			return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(d), d)), nil
		}
	default:
		return nil, errors.New("invalid data type")
	}
}
