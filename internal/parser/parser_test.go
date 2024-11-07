package parser

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
	input := NewDomainNames()
	err := input.ParseFile(path.Join(testDataPath, "lists/1.lst"))
	require.NoError(t, err)

	expected := NewDomainNames()
	expected.ParsedNames = basicList

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserMultipleFiles(t *testing.T) {
	input := NewDomainNames()
	err := input.ParseFile(path.Join(testDataPath, "lists/1.lst"))
	require.NoError(t, err)
	err = input.ParseFile(path.Join(testDataPath, "lists/2.lst"))
	require.NoError(t, err)

	expected := NewDomainNames()
	expected.ParsedNames = []string{
		"cloudflare.com",
		"google.com",
		"hashicorp.com",
		"linked.in",
		"releases.hashicorp.com",
		"rpm.releases.hashicorp.com",
		"terraform.io",
	}

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserEmpty(t *testing.T) {
	input := NewDomainNames()
	err := input.ParseFile(path.Join(testDataPath, "lists/empty.lst"))
	require.NoError(t, err)

	expected := NewDomainNames()

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserComments(t *testing.T) {
	input := NewDomainNames()
	err := input.ParseFile(path.Join(testDataPath, "lists/1.lst"))
	require.NoError(t, err)

	expected := NewDomainNames()
	expected.ParsedNames = basicList

	require.True(t, reflect.DeepEqual(input.ParsedNames, expected.ParsedNames))
	require.True(t, reflect.DeepEqual(input.UnparsedNames, expected.UnparsedNames))
}

func TestParserInvalid(t *testing.T) {
	invalidPaths := []string{
		path.Join(testDataPath, "lists/invalid_0.lst"),
		path.Join(testDataPath, "lists/invalid_1.lst"),
	}

	input := NewDomainNames()
	err := input.ParseFile(invalidPaths[0])
	require.NoError(t, err)
	err = input.ParseFile(invalidPaths[1])
	require.NoError(t, err)

	expected := NewDomainNames()
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
	input := NewDomainNames()
	err := input.ParseFile(path.Join(testDataPath, "lists/this_file_does_not_exist.lst"))
	require.Error(t, err)
}
