package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	pkgFile       = "elm-package.json"
	ext           = ".elm"
	elmStuffDir   = "elm-stuff"
	packagesDir   = "packages"
	exactDepsFile = "exact-dependencies.json"
)

var (
	separator = fmt.Sprint(filepath.Separator)

	// ErrModuleNotFound will be returned when a module has not been found.
	ErrModuleNotFound = errors.New("pkg: module not found")
	// ErrDepsNotInstalled will be returned when the list of exact dependencies
	// cannot be found. That is, when the packages have not been installed.
	ErrDepsNotInstalled = errors.New("pkg: dependencies not installed")
)

// Package represents the elm-package.json file, that is, the package manifest.
// This contains all the package information useful for the compiler, including
// its dependencies, etc.
type Package struct {
	root              string            `json:"-"`
	Version           Version           `json:"version"`
	SourceDirectories []string          `json:"source-directories"`
	NativeModules     bool              `json:"native-modules"`
	Dependencies      Dependencies      `json:"dependencies"`
	ElmVersion        VersionRange      `json:"elm-version"`
	ExactDependencies ExactDependencies `json:"-"`
}

// Root returns the package root.
func (p *Package) Root() string {
	return p.root
}

// tryLoadExactDependencies will try to load the "exact-dependencies.json" file
// inside elm-stuff director. If it's not found, it will be assumed the deps
// have not been installed and will do nothing.
// If there is an error, that error will be returned, though.
func (p *Package) tryLoadExactDependencies() error {
	path := filepath.Join(p.root, elmStuffDir, exactDepsFile)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("pkg: can't load exact dependencies: %s", err)
	}

	defer f.Close()
	var deps ExactDependencies
	if err := json.NewDecoder(f).Decode(&deps); err != nil {
		return fmt.Errorf("pkg: can't decode exact dependencies: %s", err)
	}

	p.ExactDependencies = deps
	return nil
}

// FindModule tries to find a module with the given path (path as in
// Module.Name.Path) in all the source directories.
// If the module is not in the source directories, it will try to look for it
// on the dependencies directories.
func (p *Package) FindModule(path string) (string, error) {
	if modulePath, err := p.FindSourceModule(path); err != nil {
		return "", err
	} else if modulePath != "" {
		return modulePath, nil
	}

	return p.FindDependencyModule(path)
}

// FindSourceModule tries to find a module with the given path in all the
// source directories.
func (p *Package) FindSourceModule(path string) (string, error) {
	pathParts := strings.Split(path, ".")
	for _, dir := range p.SourceDirectories {
		if path, err := p.findModuleInDir(pathParts, dir); err != nil {
			return "", err
		} else if path != "" {
			return path, nil
		}
	}

	return "", ErrModuleNotFound
}

// FindDependencyModule will try to find a module with the given path in all
// the dependency directories.
func (p *Package) FindDependencyModule(path string) (string, error) {
	if p.ExactDependencies == nil {
		return "", ErrDepsNotInstalled
	}

	pathParts := strings.Split(path, ".")
	for dep, v := range p.ExactDependencies {
		dir := filepath.Join(elmStuffDir, packagesDir, dep, v.String())
		if path, err := p.findModuleInDir(pathParts, dir); err != nil {
			return "", err
		} else if path != "" {
			return path, nil
		}
	}

	return "", ErrModuleNotFound
}

func (p *Package) findModuleInDir(pathParts []string, dir string) (string, error) {
	var path = filepath.Join(p.root, dir)
	for i, p := range pathParts {
		if i+1 == len(pathParts) {
			p = p + ext
		}
		path = filepath.Join(path, p)

		if ok, err := exists(path); err != nil {
			return "", err
		} else if ok {
			return "", nil
		}
	}

	return path, nil
}

// Dependencies is a map between a dependency name and a version range.
type Dependencies map[string]VersionRange

// ExactDependencies is a map between a dependency and the exact verson installed.
type ExactDependencies map[string]Version

// Load will load the manifest of the package from the given path until it
// reaches the root of the filesystem.
func Load(path string) (*Package, error) {
	f, root, err := findPackageFile(path)
	if err != nil {
		return nil, fmt.Errorf("pkg: can't load pkg file from path %q: %s", path, err)
	}

	if f == nil {
		return nil, nil
	}

	defer f.Close()
	var pkg Package
	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("pkg: can't decode elm-package.json: %s", err)
	}
	pkg.root = root

	if err := pkg.tryLoadExactDependencies(); err != nil {
		return nil, err
	}

	return &pkg, nil
}

func findPackageFile(path string) (io.ReadCloser, string, error) {
	if path == separator {
		return nil, "", nil
	}

	file := filepath.Join(path, pkgFile)
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return findPackageFile(filepath.Dir(path))
	}

	return f, path, err
}

func exists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
