package log

var rules = []Rule{
	//func(names []string) (level Level, matched bool) {
	//	//log.Print(names)
	//	return Info, true
	//},
	//ContainsAllNames([]string{"reader"}, Debug),
}

type Rule func(names []string) (level Level, matched bool)

func stringSliceContains(s string, ss []string) bool {
	for _, sss := range ss {
		if s == sss {
			return true
		}
	}
	return false
}

func ContainsAllNames(all []string, level Level) Rule {
	return func(names []string) (_ Level, matched bool) {
		for _, s := range all {
			//log.Println(s, all, names)
			if !stringSliceContains(s, names) {
				return
			}
		}
		return level, true
	}
}
