package requests

var TagsEmojies map[string]string

func init() {
	TagsEmojies = make(map[string]string)
	TagsEmojies["Activism"] = "✊"
	TagsEmojies["Art"] = "🎨"
	TagsEmojies["Blockchain"] = "🔗"
	TagsEmojies["Books & blogs"] = "📚"
	TagsEmojies["Career"] = "💼"
	TagsEmojies["Collaboration"] = "🤝"
	TagsEmojies["Commerce"] = "🛒"
	TagsEmojies["Culture"] = "🎎"
	TagsEmojies["DAO"] = "🚀"
	TagsEmojies["DeFi"] = "📈"
	TagsEmojies["Design"] = "🧩"
	TagsEmojies["DIY"] = "🔨"
	TagsEmojies["Environment"] = "🌿"
	TagsEmojies["Education"] = "🎒"
	TagsEmojies["Entertainment"] = "🍿"
	TagsEmojies["Ethereum"] = "Ξ"
	TagsEmojies["Event"] = "🗓"
	TagsEmojies["Fantasy"] = "🧙‍♂️"
	TagsEmojies["Fashion"] = "🧦"
	TagsEmojies["Food"] = "🌶"
	TagsEmojies["Gaming"] = "🎮"
	TagsEmojies["Global"] = "🌍"
	TagsEmojies["Health"] = "🧠"
	TagsEmojies["Hobby"] = "📐"
	TagsEmojies["Innovation"] = "🧪"
	TagsEmojies["Language"] = "📜"
	TagsEmojies["Lifestyle"] = "✨"
	TagsEmojies["Local"] = "📍"
	TagsEmojies["Love"] = "❤️"
	TagsEmojies["Markets"] = "💎"
	TagsEmojies["Movies & TV"] = "🎞"
	TagsEmojies["Music"] = "🎶"
	TagsEmojies["News"] = "🗞"
	TagsEmojies["NFT"] = "🖼"
	TagsEmojies["Non-profit"] = "🙏"
	TagsEmojies["NSFW"] = "🍆"
	TagsEmojies["Org"] = "🏢"
	TagsEmojies["Pets"] = "🐶"
	TagsEmojies["Play"] = "🎲"
	TagsEmojies["Podcast"] = "🎙️"
	TagsEmojies["Politics"] = "🗳️"
	TagsEmojies["Product"] = "🍱"
	TagsEmojies["Psyche"] = "🍁"
	TagsEmojies["Privacy"] = "👻"
	TagsEmojies["Security"] = "🔒"
	TagsEmojies["Social"] = "☕"
	TagsEmojies["Software dev"] = "👩‍💻"
	TagsEmojies["Sports"] = "⚽️"
	TagsEmojies["Tech"] = "📱"
	TagsEmojies["Travel"] = "🗺"
	TagsEmojies["Vehicles"] = "🚕"
	TagsEmojies["Web3"] = "🌐"
}

func ValidateTags(input []string) bool {
	for _, t := range input {
		_, ok := TagsEmojies[t]
		if !ok {
			return false
		}
	}

	// False if contains duplicates. Shouldn't have happened
	return len(unique(input)) == len(input)
}

func RemoveUnknownAndDeduplicateTags(input []string) []string {
	var result []string
	for _, t := range input {
		_, ok := TagsEmojies[t]
		if ok {
			result = append(result, t)
		}
	}
	return unique(result)
}

func unique(slice []string) []string {
	uniqMap := make(map[string]struct{})
	for _, v := range slice {
		uniqMap[v] = struct{}{}
	}
	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	return uniqSlice
}
