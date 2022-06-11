package valid

var URIs = [][]byte{
	[]byte("http://"),
	[]byte("https://"),
	[]byte("ftp://"),
	[]byte("mailto:"),
}

var Paths = [][]byte{
	[]byte("/"),
	[]byte("./"),
	[]byte("../"),
}
