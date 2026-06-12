package store

import "encoding/json"

func encodeTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(tags)
	return string(b)
}

func decodeTags(raw string) []string {
	if raw == "" {
		return nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil
	}
	return tags
}

func mergeTags(existing, add []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(add))
	var out []string
	for _, tag := range append(existing, add...) {
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func removeTags(existing, remove []string) []string {
	removeSet := make(map[string]struct{}, len(remove))
	for _, tag := range remove {
		removeSet[tag] = struct{}{}
	}
	var out []string
	for _, tag := range existing {
		if _, ok := removeSet[tag]; ok {
			continue
		}
		out = append(out, tag)
	}
	return out
}

func userHasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
