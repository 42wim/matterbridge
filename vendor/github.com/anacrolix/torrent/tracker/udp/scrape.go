package udp

type ScrapeRequest []InfoHash

type ScrapeResponse []ScrapeInfohashResult

type ScrapeInfohashResult struct {
	Seeders   int32
	Completed int32
	Leechers  int32
}
