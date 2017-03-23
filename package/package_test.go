package pkg

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var validPackageEntries = []entry{
	{
		"elm-package.json",
		Package{
			SourceDirectories: []string{"src", "src2"},
			Dependencies: Dependencies{
				"foo/bar": VersionRange{
					Min: Version{1, 0, 0},
					Max: Version{2, 0, 0},
				},
				"foo/baz": VersionRange{
					Min: Version{1, 0, 0},
					Max: Version{2, 0, 0},
				},
			},
		},
	},
	{"src/Foo.elm", nil},
	{"src/Foo/Bar.elm", nil},
	{"src/Foo/Bar/Baz.elm", nil},
	{"src2/Bar.elm", nil},
	{
		"elm-stuff/exact-dependencies.json",
		ExactDependencies{
			"foo/bar": Version{1, 0, 0},
			"foo/baz": Version{1, 5, 0},
		},
	},
	{
		"elm-stuff/packages/foo/bar/1.0.0/elm-package.json",
		Package{
			SourceDirectories: []string{"src"},
		},
	},
	{"elm-stuff/packages/foo/bar/1.0.0/src/Foo/Bar.elm", nil},
	{"elm-stuff/packages/foo/bar/1.0.0/src/Foo/Bar/Baz/Qux.elm", nil},
	{
		"elm-stuff/packages/foo/baz/1.5.0/elm-package.json",
		Package{
			SourceDirectories: []string{"src"},
		},
	},
	{"elm-stuff/packages/foo/baz/1.5.0/src/Foo/Bar/Baz/Mux.elm", nil},
}

var notInstalledPackageEntries = []entry{
	{
		"elm-package.json",
		Package{
			SourceDirectories: []string{"src"},
			Dependencies: Dependencies{
				"foo/bar": VersionRange{
					Min: Version{1, 0, 0},
					Max: Version{2, 0, 0},
				},
			},
		},
	},
	{"src/Foo.elm", nil},
}

var badlyInstalledPackageEntries = []entry{
	{
		"elm-package.json",
		Package{
			SourceDirectories: []string{"src"},
			Dependencies: Dependencies{
				"foo/bar": VersionRange{
					Min: Version{1, 0, 0},
					Max: Version{2, 0, 0},
				},
			},
		},
	},
	{"src/Foo.elm", nil},
	{"elm-stuff/exact-dependencies.json", "not json"},
}

func TestLoad(t *testing.T) {
	cases := []struct {
		entries []entry
		ok      bool
	}{
		{validPackageEntries, true},
		{nil, false},
		{[]entry{{"src/Foo.elm", nil}}, false},
		{[]entry{{"elm-package.json", "not json"}}, false},
		{badlyInstalledPackageEntries, false},
		{notInstalledPackageEntries, true},
	}

	require := require.New(t)
	for _, c := range cases {
		root, err := createStructure(c.entries...)
		require.NoError(err)

		pkg, err := Load(root)
		if c.ok {
			require.NotNil(pkg)
			require.NoError(err)
		} else {
			require.Error(err)
		}
	}
	createStructure(validPackageEntries...)
}

func TestFindModule(t *testing.T) {
	require := require.New(t)
	root, err := createStructure(validPackageEntries...)
	require.NoError(err)

	pkg, err := Load(root)
	require.NoError(err)

	cases := []struct {
		module   string
		expected string
		err      error
	}{
		{"Foo", "src/Foo.elm", nil},
		{"Foo", "src/Foo.elm", nil}, // this one is cached
		{"Bar", "src2/Bar.elm", nil},
		{"Foo.Bar", "src/Foo/Bar.elm", nil},
		{"Foo.Bar.Baz", "src/Foo/Bar/Baz.elm", nil},
		{"Foo.Bar.Baz.Qux", "elm-stuff/packages/foo/bar/1.0.0/src/Foo/Bar/Baz/Qux.elm", nil},
		{"Foo.Bar.Baz.Mux", "elm-stuff/packages/foo/baz/1.5.0/src/Foo/Bar/Baz/Mux.elm", nil},
		{"Foo.Bar.Baz.Mux", "elm-stuff/packages/foo/baz/1.5.0/src/Foo/Bar/Baz/Mux.elm", nil}, // this one is cached
		{"Bar.Foo", "", ErrModuleNotFound},
	}

	for _, c := range cases {
		path, err := pkg.FindModule(c.module)
		if c.err != nil {
			require.Equal(c.err, err, c.module)
		} else {
			require.NoError(err, c.module)
			require.Equal(
				filepath.Join(pkg.Root(), c.expected),
				path,
				c.module,
			)
		}
	}
}

type entry struct {
	file    string
	content interface{}
}

func createStructure(entries ...entry) (string, error) {
	root, err := ioutil.TempDir("", "testing")
	if err != nil {
		return "", err
	}

	for _, e := range entries {
		file := filepath.Join(root, e.file)
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0777); err != nil {
			return "", err
		}

		var content []byte
		switch c := e.content.(type) {
		case string:
			content = []byte(c)
		case nil:
		default:
			var err error
			content, err = json.Marshal(c)
			if err != nil {
				return "", err
			}
		}

		if err := ioutil.WriteFile(file, content, 0777); err != nil {
			return "", err
		}
	}

	return root, nil
}
