package ic

import "log"

// Convert convert bytes from CJK or UTF-8 to UTF-8 or CJK
func Convert(from string, to string, src []byte) []byte {
	if to == "utf-8" {
		out, e := ToUTF8(from, src)
		if e == nil {
			return out
		}
		log.Printf("converting from %s to UTF-8 failed: %v", from, e)
		return src
	}

	if from == "utf-8" {
		out, e := FromUTF8(to, src)
		if e == nil {
			return out
		}
		log.Printf("converting from UTF-8 to %s failed: %v", to, e)
		return src
	}
	log.Println("only converting between CJK encodings and UTF-8 is supported")
	return src
}

// ConvertString convert string from CJK or UTF-8 to UTF-8 or CJK
func ConvertString(from string, to string, src string) string {
	return string(Convert(from, to, []byte(src)))
}
