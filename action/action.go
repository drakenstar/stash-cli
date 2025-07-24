package action

type Action struct {
	Name      string
	Arguments map[string]string
}

func Parse(input string) (Action, error) {
	return Action{}, nil
}
