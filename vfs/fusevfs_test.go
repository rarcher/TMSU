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

// +build !windows

package vfs

import (
	"github.com/oniony/TMSU/entities"
	"testing"
)

func TestImplyingTagListedOnlyWithinImpliedTag(test *testing.T) {
	// set-up

	implications := entities.Implications{implication(1, "mp3", 0, "", 2, "music", 0, "")}

	// test

	topListed := isTagListed("mp3", pathToTagNames([]string{}, implications), implications)
	musicListed := isTagListed("mp3", pathToTagNames([]string{"music"}, implications), implications)
	impliedListed := isTagListed("music", pathToTagNames([]string{}, implications), implications)

	// validate

	if topListed {
		test.Fatalf("Implying tag was listed outside of the implied tag directory")
	}
	if !musicListed {
		test.Fatalf("Implying tag was not listed within the implied tag directory")
	}
	if !impliedListed {
		test.Fatalf("Implied tag was not listed")
	}
}

func TestImplyingTagListedOnlyWithinAllImpliedTags(test *testing.T) {
	// set-up

	implications := entities.Implications{implication(1, "mp3", 0, "", 2, "music", 0, ""),
		implication(1, "mp3", 0, "", 3, "audio", 0, "")}

	// test

	musicListed := isTagListed("mp3", pathToTagNames([]string{"music"}, implications), implications)
	bothListed := isTagListed("mp3", pathToTagNames([]string{"audio", "music"}, implications), implications)

	// validate

	if musicListed {
		test.Fatalf("Implying tag was listed outside of an implied tag directory")
	}
	if !bothListed {
		test.Fatalf("Implying tag was not listed within the implied tag directories")
	}
}

func TestImplyingTagListedWithinTransitivelyImpliedTag(test *testing.T) {
	// set-up

	implications := entities.Implications{implication(1, "mp3", 0, "", 2, "music", 0, ""),
		implication(2, "music", 0, "", 3, "media", 0, "")}

	// test

	mediaListed := isTagListed("mp3", pathToTagNames([]string{"media"}, implications), implications)
	musicListed := isTagListed("mp3", pathToTagNames([]string{"music"}, implications), implications)
	mediaMusicListed := isTagListed("mp3", pathToTagNames([]string{"media", "music"}, implications), implications)

	// validate

	if mediaListed {
		test.Fatalf("Implying tag was listed outside of the directly implied tag directory")
	}
	if !musicListed {
		test.Fatalf("Implying tag was not listed within the directly implied tag directory")
	}
	if !mediaMusicListed {
		test.Fatalf("Implying tag was not listed within the nested implied tag directories")
	}
}

func TestImplyingTagListedWithinImpliedTagValue(test *testing.T) {
	// set-up

	implications := entities.Implications{implication(1, "aubergine", 0, "", 2, "aka", 1, "eggplant")}

	// test

	akaListed := isTagListed("aubergine", pathToTagNames([]string{"aka"}, implications), implications)
	valueListed := isTagListed("aubergine", pathToTagNames([]string{"aka", "=eggplant"}, implications), implications)

	// validate

	if !akaListed {
		test.Fatalf("Implying tag was not listed within the implied tag directory")
	}
	if !valueListed {
		test.Fatalf("Implying tag was not listed within the implied tag value directory")
	}
}

func TestImplyingTagValueListedEverywhere(test *testing.T) {
	// set-up

	implications := entities.Implications{implication(1, "year", 1, "1990", 2, "nineties", 0, "")}

	// test

	topListed := isTagListed("year", pathToTagNames([]string{}, implications), implications)

	// validate

	if !topListed {
		test.Fatalf("Tag was not listed despite only some of its values implying other tags")
	}
}

func TestImplicationsFromPathTagValues(test *testing.T) {
	// set-up

	implications := entities.Implications{implication(1, "year", 1, "1990", 2, "nineties", 0, ""),
		implication(3, "cd", 0, "", 2, "nineties", 0, "")}

	// test

	yearListed := isTagListed("cd", pathToTagNames([]string{"year"}, implications), implications)
	valueListed := isTagListed("cd", pathToTagNames([]string{"year", "=1990"}, implications), implications)

	// validate

	if yearListed {
		test.Fatalf("Implication of a tag value was applied to a directory without that value")
	}
	if !valueListed {
		test.Fatalf("Implication of a tag value was not applied to a directory with that value")
	}
}

// unexported

func implication(tagId entities.TagId, tagName string, valueId entities.ValueId, valueName string,
	impliedTagId entities.TagId, impliedTagName string, impliedValueId entities.ValueId, impliedValueName string) *entities.Implication {
	return &entities.Implication{entities.Tag{tagId, tagName},
		entities.Value{valueId, valueName},
		entities.Tag{impliedTagId, impliedTagName},
		entities.Value{impliedValueId, impliedValueName}}
}
