package gumble

// Version represents a Mumble client or server version.
type Version struct {
	// The semantic version information as a single unsigned integer.
	//
	// Bits 0-15 are the major version, bits 16-23 are the minor version, and
	// bits 24-31 are the patch version.
	Version uint32
	// The name of the client.
	Release string
	// The operating system name.
	OS string
	// The operating system version.
	OSVersion string
}

// SemanticVersion returns the version's semantic version components.
func (v *Version) SemanticVersion() (major uint16, minor, patch uint8) {
	major = uint16(v.Version>>16) & 0xFFFF
	minor = uint8(v.Version>>8) & 0xFF
	patch = uint8(v.Version) & 0xFF
	return
}
