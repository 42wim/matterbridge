// Copyright (C) 2016 The Syncthing Authors.
//
// Adapted from https://github.com/jackpal/Taipei-Torrent/blob/dd88a8bfac6431c01d959ce3c745e74b8a911793/IGD.go
// Copyright (c) 2010 Jack Palevich (https://github.com/jackpal/Taipei-Torrent/blob/dd88a8bfac6431c01d959ce3c745e74b8a911793/LICENSE)
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

package upnp

import (
	"encoding/xml"
	"fmt"
	"net"
	"time"
)

// An IGDService is a specific service provided by an IGD.
type IGDService struct {
	UUID      string
	Device    upnpDevice
	ServiceID string
	URL       string
	URN       string
	LocalIP   net.IP
	ll        levelLogger
}

// AddPortMapping adds a port mapping to the specified IGD service.
func (s *IGDService) AddPortMapping(protocol Protocol, internalPort, externalPort int, description string, duration time.Duration) (int, error) {
	tpl := `<u:AddPortMapping xmlns:u="%s">
	<NewRemoteHost></NewRemoteHost>
	<NewExternalPort>%d</NewExternalPort>
	<NewProtocol>%s</NewProtocol>
	<NewInternalPort>%d</NewInternalPort>
	<NewInternalClient>%s</NewInternalClient>
	<NewEnabled>1</NewEnabled>
	<NewPortMappingDescription>%s</NewPortMappingDescription>
	<NewLeaseDuration>%d</NewLeaseDuration>
	</u:AddPortMapping>`
	body := fmt.Sprintf(tpl, s.URN, externalPort, protocol, internalPort, s.LocalIP, description, duration/time.Second)

	response, err := soapRequest(s.URL, s.URN, "AddPortMapping", body, s.ll)
	if err != nil && duration > 0 {
		// Try to repair error code 725 - OnlyPermanentLeasesSupported
		envelope := &soapErrorResponse{}
		if unmarshalErr := xml.Unmarshal(response, envelope); unmarshalErr != nil {
			return externalPort, unmarshalErr
		}
		if envelope.ErrorCode == 725 {
			return s.AddPortMapping(protocol, internalPort, externalPort, description, 0)
		}
	}

	return externalPort, err
}

// DeletePortMapping deletes a port mapping from the specified IGD service.
func (s *IGDService) DeletePortMapping(protocol Protocol, externalPort int) error {
	tpl := `<u:DeletePortMapping xmlns:u="%s">
	<NewRemoteHost></NewRemoteHost>
	<NewExternalPort>%d</NewExternalPort>
	<NewProtocol>%s</NewProtocol>
	</u:DeletePortMapping>`
	body := fmt.Sprintf(tpl, s.URN, externalPort, protocol)

	_, err := soapRequest(s.URL, s.URN, "DeletePortMapping", body, s.ll)
	return err
}

// GetExternalIPAddress queries the IGD service for its external IP address.
// Returns nil if the external IP address is invalid or undefined, along with
// any relevant errors
func (s *IGDService) GetExternalIPAddress() (net.IP, error) {
	tpl := `<u:GetExternalIPAddress xmlns:u="%s" />`

	body := fmt.Sprintf(tpl, s.URN)

	response, err := soapRequest(s.URL, s.URN, "GetExternalIPAddress", body, s.ll)

	if err != nil {
		return nil, err
	}

	envelope := &soapGetExternalIPAddressResponseEnvelope{}
	err = xml.Unmarshal(response, envelope)
	if err != nil {
		return nil, err
	}

	result := net.ParseIP(envelope.Body.GetExternalIPAddressResponse.NewExternalIPAddress)

	return result, nil
}

// GetLocalIPAddress returns local IP address used to contact this service
func (s *IGDService) GetLocalIPAddress() net.IP {
	return s.LocalIP
}

// ID returns a unique ID for the servic
func (s *IGDService) ID() string {
	return s.UUID + "/" + s.Device.FriendlyName + "/" + s.ServiceID + "/" + s.URN + "/" + s.URL
}
