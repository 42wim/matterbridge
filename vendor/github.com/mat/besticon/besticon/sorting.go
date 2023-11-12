package besticon

import "sort"

func sortIcons(icons []Icon, sizeDescending bool) {
	// Order after sorting: (width/height, bytes, url)
	sort.Stable(byURL(icons))
	sort.Stable(byBytes(icons))

	if sizeDescending {
		sort.Stable(sort.Reverse(byWidthHeight(icons)))
	} else {
		sort.Stable(byWidthHeight(icons))
	}
}

type byWidthHeight []Icon

func (a byWidthHeight) Len() int      { return len(a) }
func (a byWidthHeight) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byWidthHeight) Less(i, j int) bool {
	return (a[i].Width < a[j].Width) || (a[i].Height < a[j].Height)
}

type byBytes []Icon

func (a byBytes) Len() int           { return len(a) }
func (a byBytes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byBytes) Less(i, j int) bool { return (a[i].Bytes < a[j].Bytes) }

type byURL []Icon

func (a byURL) Len() int           { return len(a) }
func (a byURL) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byURL) Less(i, j int) bool { return (a[i].URL < a[j].URL) }
