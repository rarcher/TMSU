// Copyright 2011-2018 Paul Ruane.

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

package entities

import (
	"fmt"
	"github.com/oniony/TMSU/common/fingerprint"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileId uint

type FileIds []FileId

func (fileIds FileIds) Len() int {
	return len(fileIds)
}

func (fileIds FileIds) Less(i, j int) bool {
	return fileIds[i] < fileIds[j]
}

func (fileIds FileIds) Swap(i, j int) {
	fileIds[i], fileIds[j] = fileIds[j], fileIds[i]
}

func (fileIds FileIds) Uniq() FileIds {
	if len(fileIds) == 0 {
		return fileIds
	}

	sort.Sort(fileIds)
	uniq := FileIds{fileIds[0]}
	previous := fileIds[0]

	for _, fileId := range fileIds[1:] {
		if fileId != previous {
			uniq = append(uniq, fileId)
			previous = fileId
		}
	}

	return uniq
}

type File struct {
	Id          FileId
	Directory   string
	Name        string
	Fingerprint fingerprint.Fingerprint
	ModTime     time.Time
	Size        int64
	IsDir       bool
	NameMap     string
}

func (file File) Path() string {
	return filepath.Join(file.Directory, file.Name)
}

type Files []*File

func (files Files) Where(predicate func(*File) bool) Files {
	result := make(Files, 0, 10)

	for _, file := range files {
		if predicate(file) {
			result = append(result, file)
		}
	}

	return result
}

type FileTagCount struct {
	FileId    FileId
	Directory string
	Name      string
	TagCount  uint
}

func ValidateFileNameMap(nameMap string) error {
	switch nameMap {
	case "":
		return fmt.Errorf("name map cannot be empty")
	case ".", "..":
		return fmt.Errorf("name map cannot be '.' or '..'") // cannot be used in the VFS
	case "files":
		return fmt.Errorf("name map cannot be 'files'") // used in the VFS
	}

	if len(nameMap) > 255 {
		return fmt.Errorf("name map cannot be longer than 255 bytes") // cannot be used in the VFS
	}

	for _, ch := range nameMap {
		switch ch {
		case '/':
			return fmt.Errorf("name map cannot contain '%c'", ch) // cannot be used in the VFS
		case 0:
			return fmt.Errorf("name map cannot contain %U", ch) // cannot be used in the VFS
		}
	}

	// the VFS names the links to files without a name map NAME.ID.EXTENSION,
	// or NAME.ID if the file has no extension
	parts := strings.Split(nameMap, ".")
	count := len(parts)
	if (count >= 2 && isNumber(parts[count-1])) || (count >= 3 && isNumber(parts[count-2])) {
		return fmt.Errorf("name map cannot end with '.NUMBER' or '.NUMBER.EXTENSION' as these names are used for files without a name map")
	}

	return nil
}

// unexported

func isNumber(text string) bool {
	if text == "" {
		return false
	}

	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}
