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
	separator = fmt.Sprintf("%c", filepath.Separator)

	// ErrModuleNotFound will be returned when a module has not been found.
	ErrModuleNotFound = errors.New("pkg: module not found")
	// ErrDepsNotInstalled will be returned when the list of exact dependencies
	// cannot be found. That is, when the packages have not been installed.
	ErrDepsNotInstalled = errors.New("pkg: dependencies not installed")
	// ErrNotElmPackage will be returned when the given path or none of its
	// ancestors are elm packages.
	ErrNotElmPackage = errors.New("pkg: could not find an elm package in the given path or its ancestors")
)

// Package represents the elm-package.json file, that is, the package manifest.
// This contains all the package information useful for the compiler, including
// its dependencies, etc.
type Package struct {
	Version           Version           `json:"version"`
	SourceDirectories []string          `json:"source-directories"`
	NativeModules     bool              `json:"native-modules"`
	Dependencies      Dependencies      `json:"dependencies"`
	ElmVersion        VersionRange      `json:"elm-version"`
	ExactDependencies ExactDependencies `json:"-"`

	// root of the package, that is, the directory where elm-package.json is
	root string
	// dependencyCache keeps a reference to the package manifest of the
	// dependencies so we don't have to load it every time that we're looking
	// for a module
	dependencyCache map[string]*Package
	// moduleCache keeps the resolved paths for modules so they don't have to
	// looked up again
	moduleCache map[string]string
}

// Root returns the package root.
func (p *Package) Root() string {
	return p.root
}

func (p *Package) cacheModule(module string, filePath string) {
	p.moduleCache[module] = filePath
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
	if filePath, err := p.FindSourceModule(path); err != nil && err != ErrModuleNotFound {
		return "", err
	} else if filePath != "" {
		return filePath, nil
	}

	return p.FindDependencyModule(path)
}

// FindSourceModule tries to find a module with the given path in all the
// source directories.
func (p *Package) FindSourceModule(path string) (string, error) {
	if cachedPath, ok := p.moduleCache[path]; ok {
		return cachedPath, nil
	}

	pathParts := strings.Split(path, ".")
	for _, dir := range p.SourceDirectories {
		if moduleFilePath, err := p.findModuleInDir(pathParts, dir); err != nil {
			return "", err
		} else if moduleFilePath != "" {
			p.cacheModule(path, moduleFilePath)
			return moduleFilePath, nil
		}
	}

	return "", ErrModuleNotFound
}

// FindDependencyModule will try to find a module with the given path in all
// the dependency directories.
func (p *Package) FindDependencyModule(path string) (string, error) {
	if cachedPath, ok := p.moduleCache[path]; ok {
		return cachedPath, nil
	}

	if p.ExactDependencies == nil {
		return "", ErrDepsNotInstalled
	}

	for dep, v := range p.ExactDependencies {
		dir := filepath.Join(p.root, elmStuffDir, packagesDir, dep, v.String())
		var (
			pkg *Package
			ok  bool
		)

		if pkg, ok = p.dependencyCache[dep]; !ok {
			var err error
			pkg, err = loadPackage(dir, false)
			if err != nil {
				return "", fmt.Errorf("pkg: expected %s version %s to be a valid Elm package: %s", dep, v, err)
			}
		}

		moduleFilePath, err := pkg.FindSourceModule(path)
		if err != nil && err != ErrModuleNotFound {
			return "", err
		} else if moduleFilePath != "" {
			p.cacheModule(path, moduleFilePath)
			return moduleFilePath, nil
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
		} else if !ok {
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
	pkg, err := loadPackage(path, true)
	if err != nil {
		return nil, err
	}

	pkg.dependencyCache = make(map[string]*Package)

	if err := pkg.tryLoadExactDependencies(); err != nil {
		return nil, err
	}

	return pkg, nil
}

func loadPackage(path string, recursive bool) (*Package, error) {
	f, root, err := findPackageFile(path, recursive)
	if err != nil {
		return nil, fmt.Errorf("pkg: can't load pkg file from path %q: %s", path, err)
	}

	if f == nil {
		return nil, ErrNotElmPackage
	}

	defer f.Close()
	var pkg Package
	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("pkg: can't decode elm-package.json: %s", err)
	}
	pkg.root = root
	pkg.moduleCache = make(map[string]string)
	return &pkg, nil
}

func findPackageFile(path string, recursive bool) (io.ReadCloser, string, error) {
	if path == separator {
		return nil, "", nil
	}

	file := filepath.Join(path, pkgFile)
	f, err := os.Open(file)
	if os.IsNotExist(err) && recursive {
		return findPackageFile(filepath.Dir(path), true)
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
