package deprecation

const (
	ChatProfileDeprecated  = true
	ChatTimelineDeprecated = true
)

func AddChatsCount(count int) int {
	var add = 0
	if !ChatProfileDeprecated {
		add++
	}
	if !ChatTimelineDeprecated {
		add++
	}
	return count + add
}

func AddProfileFiltersCount(count int) int {
	if ChatProfileDeprecated {
		return count
	}
	return count + 1
}

func AddTimelineFiltersCount(count int) int {
	if ChatTimelineDeprecated {
		return count
	}
	return count + 1
}
