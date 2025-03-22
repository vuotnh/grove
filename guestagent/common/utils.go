package guestagent

import (
	"encoding/json"
	"errors"
	"fmt"
	"grove/common"
	"os"
	"strings"
)

type StreamCodec interface {
	Serialize(data interface{}) ([]byte, error)
	Deserialize(data []byte, result interface{}) error
}

type JsonCodec struct{}

func (j JsonCodec) Serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (j JsonCodec) Deserialize(data []byte, result interface{}) error {
	return json.Unmarshal(data, result)
}

func ReadFile(path string, codec StreamCodec, decode bool) (interface{}, error) {
	// Kiểm tra file có tồn tại hay không
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("File does not exist: " + path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if decode {
		var result interface{}
		err := codec.Deserialize(data, &result)
		return result, err
	}

	return data, nil
}

func ExistsPath(path string, isDirectory bool, asRoot bool) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	found := (isDirectory && info.IsDir()) || (isDirectory && !info.IsDir())

	// Only check as root if we can't see it as the regular user, since
	if !found && asRoot {
		testFlag := "-f"
		if isDirectory {
			testFlag = "-d"
		}
		cmd := fmt.Sprintf("test %s %s && echo 1 || echo 0", testFlag, path)
		args := strings.Split(cmd, " ")
		kwargs := map[string]interface{}{
			"log_output_on_error": true,
		}
		stdout, _ := common.ExecuteWithTimeout(args, kwargs)
		if stdout != "" {
			return true
		} else {
			return false
		}
	}
	return found
}
