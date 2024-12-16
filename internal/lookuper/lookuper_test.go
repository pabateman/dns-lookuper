package lookuper

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	outputDirectory          = "../../testdata/lookuper/output"
	listsDirectory           = "../../testdata/lookuper/lists"
	expectedContentDirectory = "../../testdata/lookuper/expected"
)

func getStringFromFile(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error while opening file: %v", err)
	}

	return string(file), nil
}

func TestTaskBasic(t *testing.T) {
	output, err := os.CreateTemp(outputDirectory, "lookuper-output")
	require.Nil(t, err)

	defer output.Close()
	defer os.Remove(output.Name())

	settings := &settings{
		LookupTimeout: "15s",
		Fail:          false,
	}

	task := &task{
		Files: []string{
			path.Join(listsDirectory, "basic.lst"),
		},
		Output: output.Name(),
	}

	err = performTask(task, settings)
	require.Nil(t, err)

	expected, err := getStringFromFile(path.Join(expectedContentDirectory, "basic.txt"))
	require.Nil(t, err)

	actual, err := getStringFromFile(output.Name())
	require.Nil(t, err)

	require.Equal(t, expected, actual)
}
