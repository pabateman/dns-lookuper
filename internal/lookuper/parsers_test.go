package lookuper

import (
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testDataPath = "../../testdata"
	basicList    = []string{"cloudflare.com", "hashicorp.com", "terraform.io"}
)

func TestParserBasic(t *testing.T) {
	input := newDomainNames()
	err := input.parseFile(path.Join(testDataPath, "lists/1.lst"))
	require.NoError(t, err)

	expected := newDomainNames()
	expected.ParsedNames = basicList

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserEmpty(t *testing.T) {
	input := newDomainNames()
	err := input.parseFile(path.Join(testDataPath, "lists/empty.lst"))
	require.NoError(t, err)

	expected := newDomainNames()

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserComments(t *testing.T) {
	input := newDomainNames()
	err := input.parseFile(path.Join(testDataPath, "lists/1.lst"))
	require.NoError(t, err)

	expected := newDomainNames()
	expected.ParsedNames = basicList

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserInvalid(t *testing.T) {
	invalidPaths := []string{
		path.Join(testDataPath, "lists/invalid_0.lst"),
		path.Join(testDataPath, "lists/invalid_1.lst"),
	}

	input := newDomainNames()
	err := input.parseFile(invalidPaths[0])
	require.NoError(t, err)
	err = input.parseFile(invalidPaths[1])
	require.NoError(t, err)

	expected := newDomainNames()
	expected.ParsedNames = []string{"cncf.io", "docker.com", "github.com", "iana.org", "stackoverflow.com", "stackstatus.net"}
	expected.UnparsedNames[invalidPaths[0]] = []string{
		"invalid$name",
		"another/invalid/name",
	}
	expected.UnparsedNames[invalidPaths[1]] = []string{
		"help/stackstatus.net",
	}

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserNonExistingFile(t *testing.T) {
	input := newDomainNames()
	err := input.parseFile(path.Join(testDataPath, "lists/this_file_does_not_exist.lst"))
	require.Error(t, err)
}
