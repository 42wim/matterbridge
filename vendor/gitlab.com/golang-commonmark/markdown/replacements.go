// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import "strings"

func exclquest(b byte) bool {
	return b == '!' || b == '?'
}

func byteToLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b - 'A' + 'a'
	}
	return b
}

var replChar = [256]bool{
	'(': true,
	'!': true,
	'+': true,
	',': true,
	'-': true,
	'.': true,
	'?': true,
}

func performReplacements(s string) string {
	var ss []string

	start := 0
	for i := 0; i < len(s); i++ {
		b := s[i]

		if replChar[b] {

		outer:
			switch b {
			case '(':
				if i+2 >= len(s) {
					break
				}

				b2 := s[i+1]

				b2 = byteToLower(b2)
				switch b2 {
				case 'c', 'r', 'p':
					if s[i+2] != ')' {
						break outer
					}
					switch b2 {
					case 'c':
						if start < i {
							ss = append(ss, s[start:i])
						}
						ss = append(ss, "©")
					case 'r':
						if start < i {
							ss = append(ss, s[start:i])
						}
						ss = append(ss, "®")
					case 'p':
						if start < i {
							ss = append(ss, s[start:i])
						}
						ss = append(ss, "§")
					}
					i += 2
					start = i + 1
					continue

				case 't':
					if i+3 >= len(s) {
						break outer
					}
					if s[i+3] != ')' || byteToLower(s[i+2]) != 'm' {
						break outer
					}
					if start < i {
						ss = append(ss, s[start:i])
					}
					ss = append(ss, "™")
					i += 3
					start = i + 1
					continue
				default:
					break outer
				}

			case '+':
				if i+1 >= len(s) || s[i+1] != '-' {
					break
				}
				if start < i {
					ss = append(ss, s[start:i])
				}
				ss = append(ss, "±")
				i++
				start = i + 1
				continue

			case '.':
				if i+1 >= len(s) || s[i+1] != '.' {
					break
				}

				j := i + 2
				for j < len(s) && s[j] == '.' {
					j++
				}
				if start < i {
					ss = append(ss, s[start:i])
				}
				if i == 0 || !(s[i-1] == '?' || s[i-1] == '!') {
					ss = append(ss, "…")
				} else {
					ss = append(ss, "..")
				}
				i = j - 1
				start = i + 1
				continue

			case '?', '!':
				if i+3 >= len(s) {
					break
				}
				if !(exclquest(s[i+1]) && exclquest(s[i+2]) && exclquest(s[i+3])) {
					break
				}
				if start < i {
					ss = append(ss, s[start:i])
				}
				ss = append(ss, s[i:i+3])
				j := i + 3
				for j < len(s) && exclquest(s[j]) {
					j++
				}
				i = j - 1
				start = i + 1
				continue

			case ',':
				if i+1 >= len(s) || s[i+1] != ',' {
					break
				}
				if start < i {
					ss = append(ss, s[start:i])
				}
				ss = append(ss, ",")
				j := i + 2
				for j < len(s) && s[j] == ',' {
					j++
				}
				i = j - 1
				start = i + 1
				continue

			case '-':
				if i+1 >= len(s) || s[i+1] != '-' {
					break
				}
				if i+2 >= len(s) || s[i+2] != '-' {
					if start < i {
						ss = append(ss, s[start:i])
					}
					ss = append(ss, "–")
					i++
					start = i + 1
					continue
				}
				if i+3 >= len(s) || s[i+3] != '-' {
					if start < i {
						ss = append(ss, s[start:i])
					}
					ss = append(ss, "—")
					i += 2
					start = i + 1
					continue
				}

				j := i + 3
				for j < len(s) && s[j] == '-' {
					j++
				}
				if start < i {
					ss = append(ss, s[start:i])
				}
				ss = append(ss, s[i:j])
				i = j - 1
				start = i + 1
				continue
			}
		}
	}
	if ss == nil {
		return s
	}
	if start < len(s) {
		ss = append(ss, s[start:])
	}
	return strings.Join(ss, "")
}

func ruleReplacements(s *StateCore) {
	if !s.Md.Typographer {
		return
	}

	insideLink := false
	for _, tok := range s.Tokens {
		if tok, ok := tok.(*Inline); ok {
			for _, itok := range tok.Children {
				switch itok := itok.(type) {
				case *LinkOpen:
					insideLink = true
				case *LinkClose:
					insideLink = false
				case *Text:
					if !insideLink {
						itok.Content = performReplacements(itok.Content)
					}
				}
			}
		}
	}
}
