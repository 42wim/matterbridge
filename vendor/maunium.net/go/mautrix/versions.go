// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mautrix

import (
	"fmt"
	"regexp"
	"strconv"
)

// RespVersions is the JSON response for https://spec.matrix.org/v1.2/client-server-api/#get_matrixclientversions
type RespVersions struct {
	Versions         []SpecVersion   `json:"versions"`
	UnstableFeatures map[string]bool `json:"unstable_features"`
}

func (versions *RespVersions) ContainsFunc(match func(found SpecVersion) bool) bool {
	for _, found := range versions.Versions {
		if match(found) {
			return true
		}
	}
	return false
}

func (versions *RespVersions) Contains(version SpecVersion) bool {
	return versions.ContainsFunc(func(found SpecVersion) bool {
		return found == version
	})
}

func (versions *RespVersions) ContainsGreaterOrEqual(version SpecVersion) bool {
	return versions.ContainsFunc(func(found SpecVersion) bool {
		return found.GreaterThan(version) || found == version
	})
}

func (versions *RespVersions) GetLatest() (latest SpecVersion) {
	for _, ver := range versions.Versions {
		if ver.GreaterThan(latest) {
			latest = ver
		}
	}
	return
}

type SpecVersionFormat int

const (
	SpecVersionFormatUnknown SpecVersionFormat = iota
	SpecVersionFormatR
	SpecVersionFormatV
)

var (
	SpecR000 = MustParseSpecVersion("r0.0.0")
	SpecR001 = MustParseSpecVersion("r0.0.1")
	SpecR010 = MustParseSpecVersion("r0.1.0")
	SpecR020 = MustParseSpecVersion("r0.2.0")
	SpecR030 = MustParseSpecVersion("r0.3.0")
	SpecR040 = MustParseSpecVersion("r0.4.0")
	SpecR050 = MustParseSpecVersion("r0.5.0")
	SpecR060 = MustParseSpecVersion("r0.6.0")
	SpecR061 = MustParseSpecVersion("r0.6.1")
	SpecV11  = MustParseSpecVersion("v1.1")
	SpecV12  = MustParseSpecVersion("v1.2")
	SpecV13  = MustParseSpecVersion("v1.3")
	SpecV14  = MustParseSpecVersion("v1.4")
	SpecV15  = MustParseSpecVersion("v1.5")
)

func (svf SpecVersionFormat) String() string {
	switch svf {
	case SpecVersionFormatR:
		return "r"
	case SpecVersionFormatV:
		return "v"
	default:
		return ""
	}
}

type SpecVersion struct {
	Format SpecVersionFormat
	Major  int
	Minor  int
	Patch  int

	Raw string
}

var legacyVersionRegex = regexp.MustCompile(`^r(\d+)\.(\d+)\.(\d+)$`)
var modernVersionRegex = regexp.MustCompile(`^v(\d+)\.(\d+)$`)

func MustParseSpecVersion(version string) SpecVersion {
	sv, err := ParseSpecVersion(version)
	if err != nil {
		panic(err)
	}
	return sv
}

func ParseSpecVersion(version string) (sv SpecVersion, err error) {
	sv.Raw = version
	if parts := modernVersionRegex.FindStringSubmatch(version); parts != nil {
		sv.Major, _ = strconv.Atoi(parts[1])
		sv.Minor, _ = strconv.Atoi(parts[2])
		sv.Format = SpecVersionFormatV
	} else if parts = legacyVersionRegex.FindStringSubmatch(version); parts != nil {
		sv.Major, _ = strconv.Atoi(parts[1])
		sv.Minor, _ = strconv.Atoi(parts[2])
		sv.Patch, _ = strconv.Atoi(parts[3])
		sv.Format = SpecVersionFormatR
	} else {
		err = fmt.Errorf("version '%s' doesn't match either known syntax", version)
	}
	return
}

func (sv *SpecVersion) UnmarshalText(version []byte) error {
	*sv, _ = ParseSpecVersion(string(version))
	return nil
}

func (sv *SpecVersion) MarshalText() ([]byte, error) {
	return []byte(sv.String()), nil
}

func (sv SpecVersion) String() string {
	switch sv.Format {
	case SpecVersionFormatR:
		return fmt.Sprintf("r%d.%d.%d", sv.Major, sv.Minor, sv.Patch)
	case SpecVersionFormatV:
		return fmt.Sprintf("v%d.%d", sv.Major, sv.Minor)
	default:
		return sv.Raw
	}
}

func (sv SpecVersion) LessThan(other SpecVersion) bool {
	return sv != other && !sv.GreaterThan(other)
}

func (sv SpecVersion) GreaterThan(other SpecVersion) bool {
	return sv.Format > other.Format ||
		(sv.Format == other.Format && sv.Major > other.Major) ||
		(sv.Format == other.Format && sv.Major == other.Major && sv.Minor > other.Minor) ||
		(sv.Format == other.Format && sv.Major == other.Major && sv.Minor == other.Minor && sv.Patch > other.Patch)
}
