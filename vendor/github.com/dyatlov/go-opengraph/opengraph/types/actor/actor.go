package actor

// Actor contain Open Graph Actor structure
type Actor struct {
	Profile string `json:"profile"`
	Role    string `json:"role"`
}

func NewActor() *Actor {
	return &Actor{}
}

func AddProfile(actors []*Actor, v string) []*Actor {
	if len(actors) == 0 || actors[len(actors)-1].Profile != "" {
		actors = append(actors, &Actor{})
	}
	actors[len(actors)-1].Profile = v
	return actors
}

func AddRole(actors []*Actor, v string) []*Actor {
	if len(actors) == 0 || actors[len(actors)-1].Role != "" {
		actors = append(actors, &Actor{})
	}
	actors[len(actors)-1].Role = v
	return actors
}
