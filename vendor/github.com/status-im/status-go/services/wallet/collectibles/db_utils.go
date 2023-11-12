package collectibles

func insertStatement(allowUpdate bool) string {
	if allowUpdate {
		return `INSERT OR REPLACE`
	}
	return `INSERT OR IGNORE`
}
