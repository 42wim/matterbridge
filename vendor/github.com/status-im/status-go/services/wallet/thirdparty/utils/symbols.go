package utils

import "strings"

var renameMapping = map[string]string{
	"STT": "SNT",
}

func RenameSymbols(symbols []string) (renames []string) {
	for _, symbol := range symbols {
		renames = append(renames, GetRealSymbol(symbol))
	}
	return
}

func GetRealSymbol(symbol string) string {
	if val, ok := renameMapping[strings.ToUpper(symbol)]; ok {
		return val
	}
	return strings.ToUpper(symbol)
}

func ChunkSymbols(symbols []string, chunkSizeOptional ...int) [][]string {
	var chunks [][]string
	chunkSize := 100
	if len(chunkSizeOptional) > 0 {
		chunkSize = chunkSizeOptional[0]
	}

	for i := 0; i < len(symbols); i += chunkSize {
		end := i + chunkSize

		if end > len(symbols) {
			end = len(symbols)
		}

		chunks = append(chunks, symbols[i:end])
	}

	return chunks
}
