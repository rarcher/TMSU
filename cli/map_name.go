// Copyright 2026 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cli

import (
	"fmt"
	"github.com/oniony/TMSU/common/log"
	_path "github.com/oniony/TMSU/common/path"
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage"
	"os"
	"path/filepath"
)

var MapNameCommand = Command{
	Name:     "map-name",
	Synopsis: "Set the name of a file in the virtual filesystem",
	Usages: []string{"tmsu map-name [OPTION]... FILE NAME",
		"tmsu map-name [OPTION]... FILE",
		"tmsu map-name [OPTION]... --clear FILE..."},
	Description: `Sets the name of the symbolic links to FILE within the virtual filesystem to NAME.

By default the symbolic links in the virtual filesystem are named after the file they point to, with the file's identifier added before the extension. When a file has a name map, NAME is used exactly as given instead.

A NAME can only be used by one file. It cannot contain a slash, cannot be longer than 255 bytes and cannot be one of the names reserved by the virtual filesystem: '.', '..' and 'files'.

NAME also cannot end with a number after a dot, whether or not an extension follows, e.g. 'mountain.7' or 'mountain.7.jpg', as the virtual filesystem uses names of this form for files without a name map.

If NAME is not specified then the name map of FILE, if it has one, is shown. The --clear option removes the name map so that the default name is used again.

FILE must already be tagged.`,
	Examples: []string{`$ tmsu map-name IMG_2041.jpg "Summit at dawn.jpg"`,
		"$ tmsu map-name IMG_2041.jpg\nSummit at dawn.jpg",
		"$ tmsu map-name --clear IMG_2041.jpg"},
	Options: Options{{"--clear", "-c", "remove the name map", false, ""},
		{"--no-dereference", "-P", "do not follow symbolic links (map the link itself)", false, ""}},
	Exec: mapNameExec,
}

// unexported

func mapNameExec(options Options, args []string, databasePath string) (error, warnings) {
	if len(args) < 1 {
		return fmt.Errorf("too few arguments"), nil
	}

	clearNameMap := options.HasOption("--clear")
	followSymlinks := !options.HasOption("--no-dereference")

	if !clearNameMap && len(args) > 2 {
		return fmt.Errorf("too many arguments"), nil
	}

	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	if clearNameMap {
		return clearNameMaps(store, tx, args, followSymlinks)
	}

	if len(args) == 1 {
		return showNameMap(store, tx, args[0], followSymlinks), nil
	}

	return mapName(store, tx, args[0], args[1], followSymlinks), nil
}

func mapName(store *storage.Storage, tx *storage.Tx, path, nameMap string, followSymlinks bool) error {
	file, err := fileForNameMap(store, tx, path, followSymlinks)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("%v: file is not tagged", path)
	}

	existingFile, err := store.FileByNameMap(tx, nameMap)
	if err != nil {
		return fmt.Errorf("could not retrieve file with name map '%v': %v", nameMap, err)
	}
	if existingFile != nil {
		if existingFile.Id == file.Id {
			return nil
		}

		return fmt.Errorf("name map '%v' is already used by '%v'", nameMap, existingFile.Path())
	}

	log.Infof(2, "%v: setting name map to '%v'.", path, nameMap)

	if err := store.UpdateFileNameMap(tx, file.Id, nameMap); err != nil {
		return fmt.Errorf("%v: could not set name map to '%v': %v", path, nameMap, err)
	}

	return nil
}

func showNameMap(store *storage.Storage, tx *storage.Tx, path string, followSymlinks bool) error {
	file, err := fileForNameMap(store, tx, path, followSymlinks)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("%v: file is not tagged", path)
	}

	if file.NameMap != "" {
		fmt.Println(file.NameMap)
	}

	return nil
}

func clearNameMaps(store *storage.Storage, tx *storage.Tx, paths []string, followSymlinks bool) (error, warnings) {
	warnings := make(warnings, 0, 10)

	for _, path := range paths {
		file, err := fileForNameMap(store, tx, path, followSymlinks)
		if err != nil {
			return err, warnings
		}
		if file == nil {
			warnings = append(warnings, fmt.Sprintf("%v: file is not tagged", path))
			continue
		}

		log.Infof(2, "%v: removing name map.", path)

		if err := store.DeleteFileNameMap(tx, file.Id); err != nil {
			return fmt.Errorf("%v: could not remove name map: %v", path, err), warnings
		}
	}

	return nil, warnings
}

func fileForNameMap(store *storage.Storage, tx *storage.Tx, path string, followSymlinks bool) (*entities.File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("%v: could not get absolute path: %v", path, err)
	}

	log.Infof(2, "%v: resolving path", path)

	stat, err := os.Lstat(absPath)
	if err != nil {
		switch {
		case os.IsNotExist(err), os.IsPermission(err):
			// ignore
		default:
			return nil, err
		}
	} else if stat.Mode()&os.ModeSymlink != 0 && followSymlinks {
		absPath, err = _path.Dereference(absPath)
		if err != nil {
			return nil, err
		}
	}

	file, err := store.FileByPath(tx, absPath)
	if err != nil {
		return nil, fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}

	return file, nil
}
