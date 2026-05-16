package app

func needsEntityResolution(inputs []string) bool {
	for _, input := range inputs {
		if input != "" && !isLikelyEntityID(input) {
			return true
		}
	}
	return false
}

func needsSingleEntityResolution(input *string) bool {
	if input == nil {
		return false
	}
	if *input == "" || *input == "current" {
		return false
	}
	return !isLikelyEntityID(*input)
}

func maybeIDs(inputs []string) []string {
	if !needsEntityResolution(inputs) {
		return inputs
	}
	return nil
}

func maybeSingleID(input *string) []string {
	if input == nil || needsSingleEntityResolution(input) || *input == "current" {
		return nil
	}
	return []string{*input}
}

func fallbackIDs(resolved []string, fallback []string) []string {
	if len(resolved) > 0 {
		return resolved
	}
	return fallback
}
