package ens

import (
	"fmt"
	"strings"
)

// DomainLevel calculates the level of the domain presented.
// A top-level domain (e.g. 'eth') will be 0, a domain (e.g.
// 'foo.eth') will be 1, a subdomain (e.g. 'bar.foo.eth' will
// be 2, etc.
func DomainLevel(name string) (level int) {
	return len(strings.Split(name, ".")) - 1
}

// NormaliseDomain turns ENS domain in to normal form
func NormaliseDomain(domain string) (string, error) {
	wildcard := false
	if strings.HasPrefix(domain, "*.") {
		wildcard = true
		domain = domain[2:]
	}
	output, err := p.ToUnicode(strings.ToLower(domain))
	if err != nil {
		return "", err
	}

	// ToUnicode() removes leading periods.  Replace them
	if strings.HasPrefix(domain, ".") && !strings.HasPrefix(output, ".") {
		output = "." + output
	}

	// If we removed a wildcard then add it back
	if wildcard {
		output = "*." + output
	}
	return output, nil
}

// NormaliseDomainStrict turns ENS domain in to normal form, using strict DNS
// rules (e.g. no underscores)
func NormaliseDomainStrict(domain string) (string, error) {
	wildcard := false
	if strings.HasPrefix(domain, "*.") {
		wildcard = true
		domain = domain[2:]
	}
	output, err := pStrict.ToUnicode(strings.ToLower(domain))
	if err != nil {
		return "", err
	}

	// ToUnicode() removes leading periods.  Replace them
	if strings.HasPrefix(domain, ".") && !strings.HasPrefix(output, ".") {
		output = "." + output
	}

	// If we removed a wildcard then add it back
	if wildcard {
		output = "*." + output
	}
	return output, nil
}

// Tld obtains the top-level domain of an ENS name
func Tld(domain string) string {
	domain, err := NormaliseDomain(domain)
	if err != nil {
		return domain
	}
	tld, err := DomainPart(domain, -1)
	if err != nil {
		return domain
	}
	return tld
}

// Domain obtains the domain of an ENS name, including subdomains.  It does this
// by removing everything up to and including the first period.
// For example, 'eth' will return ''
//              'foo.eth' will return 'eth'
//              'bar.foo.eth' will return 'foo.eth'
func Domain(domain string) string {
	if idx := strings.IndexByte(domain, '.'); idx >= 0 {
		return domain[idx+1:]
	}
	return ""
}

// DomainPart obtains a part of a name
// Positive parts start at the lowest-level of the domain and work towards the
// top-level domain.  Negative parts start at the top-level domain and work
// towards the lowest-level domain.
// For example, with a domain bar.foo.com the following parts will be returned:
// Number | part
//      1 |  bar
//      2 |  foo
//      3 |  com
//     -1 |  com
//     -2 |  foo
//     -3 |  bar
func DomainPart(domain string, part int) (string, error) {
	if part == 0 {
		return "", fmt.Errorf("invalid part")
	}
	domain, err := NormaliseDomain(domain)
	if err != nil {
		return "", err
	}
	parts := strings.Split(domain, ".")
	if len(parts) < abs(part) {
		return "", fmt.Errorf("not enough parts")
	}
	if part < 0 {
		return parts[len(parts)+part], nil
	}
	return parts[part-1], nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// UnqualifiedName strips the root from the domain and ensures the result is
// suitable as a name
func UnqualifiedName(domain string, root string) (string, error) {
	suffix := fmt.Sprintf(".%s", root)
	name := strings.TrimSuffix(domain, suffix)
	if strings.Contains(name, ".") {
		return "", fmt.Errorf("%s not a direct child of %s", domain, root)
	}
	return name, nil
}
