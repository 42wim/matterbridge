// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"

	"maunium.net/go/mautrix/util"
)

// Registration contains the data in a Matrix appservice registration.
// See https://spec.matrix.org/v1.2/application-service-api/#registration
type Registration struct {
	ID              string     `yaml:"id" json:"id"`
	URL             string     `yaml:"url" json:"url"`
	AppToken        string     `yaml:"as_token" json:"as_token"`
	ServerToken     string     `yaml:"hs_token" json:"hs_token"`
	SenderLocalpart string     `yaml:"sender_localpart" json:"sender_localpart"`
	RateLimited     *bool      `yaml:"rate_limited,omitempty" json:"rate_limited,omitempty"`
	Namespaces      Namespaces `yaml:"namespaces" json:"namespaces"`
	Protocols       []string   `yaml:"protocols,omitempty" json:"protocols,omitempty"`

	SoruEphemeralEvents bool `yaml:"de.sorunome.msc2409.push_ephemeral,omitempty" json:"de.sorunome.msc2409.push_ephemeral,omitempty"`
	EphemeralEvents     bool `yaml:"push_ephemeral,omitempty" json:"push_ephemeral,omitempty"`
}

// CreateRegistration creates a Registration with random appservice and homeserver tokens.
func CreateRegistration() *Registration {
	return &Registration{
		AppToken:    util.RandomString(64),
		ServerToken: util.RandomString(64),
	}
}

// LoadRegistration loads a YAML file and turns it into a Registration.
func LoadRegistration(path string) (*Registration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reg := &Registration{}
	err = yaml.Unmarshal(data, reg)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

// Save saves this Registration into a file at the given path.
func (reg *Registration) Save(path string) error {
	data, err := yaml.Marshal(reg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// YAML returns the registration in YAML format.
func (reg *Registration) YAML() (string, error) {
	data, err := yaml.Marshal(reg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Namespaces contains the three areas that appservices can reserve parts of.
type Namespaces struct {
	UserIDs     NamespaceList `yaml:"users,omitempty" json:"users,omitempty"`
	RoomAliases NamespaceList `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	RoomIDs     NamespaceList `yaml:"rooms,omitempty" json:"rooms,omitempty"`
}

// Namespace is a reserved namespace in any area.
type Namespace struct {
	Regex     string `yaml:"regex" json:"regex"`
	Exclusive bool   `yaml:"exclusive" json:"exclusive"`
}

type NamespaceList []Namespace

func (nsl *NamespaceList) Register(regex *regexp.Regexp, exclusive bool) {
	ns := Namespace{
		Regex:     regex.String(),
		Exclusive: exclusive,
	}
	if nsl == nil {
		*nsl = []Namespace{ns}
	} else {
		*nsl = append(*nsl, ns)
	}
}
