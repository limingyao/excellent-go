package funcx

func MapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MergeMaps[V any](maps ...map[string]V) map[string]V {
	result := make(map[string]V)
	for i := range maps {
		for k, v := range maps[i] {
			result[k] = v
		}
	}
	return result
}
