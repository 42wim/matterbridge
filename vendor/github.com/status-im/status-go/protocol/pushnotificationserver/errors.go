package pushnotificationserver

import "errors"

var ErrInvalidPushNotificationRegistrationVersion = errors.New("invalid version")
var ErrEmptyPushNotificationRegistrationPayload = errors.New("empty payload")
var ErrMalformedPushNotificationRegistrationInstallationID = errors.New("invalid installationID")
var ErrEmptyPushNotificationRegistrationPublicKey = errors.New("no public key")
var ErrCouldNotUnmarshalPushNotificationRegistration = errors.New("could not unmarshal preferences")
var ErrMalformedPushNotificationRegistrationDeviceToken = errors.New("invalid device token")
var ErrMalformedPushNotificationRegistrationGrant = errors.New("invalid grant")
var ErrMalformedPushNotificationRegistrationAccessToken = errors.New("invalid access token")
var ErrUnknownPushNotificationRegistrationTokenType = errors.New("invalid token type")
