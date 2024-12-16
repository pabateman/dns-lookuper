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

func getFilesAsString(paths ...string) ([]string, error) {
	result := make([]string, 0)

	for _, path := range paths {

		file, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error while opening file: %v", err)
		}
		result = append(result, string(file))
	}

	return result, nil
}

func TestTaskBasic(t *testing.T) {
	output := path.Join(outputDirectory, "actual-basic.txt")

	settings := &settings{
		LookupTimeout: "15s",
		Fail:          false,
	}

	task := &task{
		Files: []string{
			path.Join(listsDirectory, "basic.lst"),
		},
		Output: output,
	}

	err := performTask(task, settings)
	require.Nil(t, err)

	expected, err := getFilesAsString(path.Join(expectedContentDirectory, "basic-list.txt"))
	require.Nil(t, err)

	actual, err := getFilesAsString(output)
	require.Nil(t, err)

	require.Equal(t, expected, actual)

	err = os.Remove(output)
	require.Nil(t, err)
}

func TestTaskMultiple(t *testing.T) {
	outputs := []string{
		path.Join(outputDirectory, "actual-multiple-01.txt"),
		path.Join(outputDirectory, "actual-multiple-02.txt"),
	}

	inputs := []string{
		path.Join(listsDirectory, "basic.lst"),
		path.Join(listsDirectory, "multiple-01.lst"),
		path.Join(listsDirectory, "multiple-02.lst"),
	}

	expecteds := []string{
		path.Join(expectedContentDirectory, "multiple-01.txt"),
		path.Join(expectedContentDirectory, "multiple-02.txt"),
	}

	config := &config{
		Settings: &settings{
			LookupTimeout: "15s",
			Fail:          false,
		},

		Tasks: []task{
			{
				Files:  inputs[:2],
				Output: outputs[0],
			},
			{
				Files:  inputs[2:],
				Output: outputs[1],
			},
		},
	}

	err := walkTasks(config)
	require.Nil(t, err)

	actual, err := getFilesAsString(outputs...)
	require.Nil(t, err)

	expected, err := getFilesAsString(expecteds...)
	require.Nil(t, err)

	require.Equal(t, expected, actual)

	config.Settings.Fail = true
	err = walkTasks(config)
	require.NotNil(t, err)

	for _, output := range outputs {
		err = os.Remove(output)
		require.Nil(t, err)
	}
}

func TestTaskFails(t *testing.T) {
	output := path.Join(outputDirectory, "actual-multiple-02.txt")
	input := path.Join(listsDirectory, "multiple-02.lst")

	config := &config{
		Settings: &settings{
			LookupTimeout: "15s",
			Fail:          false,
		},

		Tasks: []task{
			{
				Files:  []string{input},
				Output: output,
			},
		},
	}

	err := walkTasks(config)
	require.Nil(t, err)

	expected, err := getFilesAsString(path.Join(expectedContentDirectory, "multiple-02.txt"))
	require.Nil(t, err)

	actual, err := getFilesAsString(output)
	require.Nil(t, err)

	require.Equal(t, expected, actual)

	config.Settings.Fail = true
	err = walkTasks(config)
	require.NotNil(t, err)

	config.Settings.Fail = false
	err = walkTasks(config)
	require.Nil(t, err)

	config.Settings.LookupTimeout = "1"
	err = walkTasks(config)
	require.NotNil(t, err)

	config.Settings.LookupTimeout = "15s"
	err = walkTasks(config)
	require.Nil(t, err)

	inputNxdomain := path.Join(listsDirectory, "inputnxdomain.lst")
	outputNxdomainPath := path.Join(outputDirectory, "outputnxdomain.lst")
	expectedNxdomainPath := path.Join(expectedContentDirectory, "stub")

	config.Tasks[0].Files = []string{inputNxdomain}
	config.Tasks[0].Output = outputNxdomainPath

	err = walkTasks(config)
	require.Nil(t, err)

	actualNxdomain, err := getFilesAsString(outputNxdomainPath)
	require.Nil(t, err)

	expectedNxdomain, err := getFilesAsString(expectedNxdomainPath)
	require.Nil(t, err)

	require.Equal(t, expectedNxdomain, actualNxdomain)

	config.Settings.Fail = true
	err = walkTasks(config)
	require.NotNil(t, err)
}
