package wray

func contains(target string, slice []string) bool {
  for _, t := range(slice) {
    if t == target {
      return true
    }
  }
  return false
}
