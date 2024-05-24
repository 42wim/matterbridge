package profile

import "strings"

// Profile contain Open Graph Profile structure
type Profile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Gender    string `json:"gender"`
}

func NewProfile() *Profile {
	return &Profile{}
}

func AddBasicProfile(profiles []*Profile, v string) []*Profile {
	parts := strings.SplitN(v, " ", 2)
	if len(profiles) == 0 || profiles[len(profiles)-1].FirstName != "" {
		profiles = append(profiles, &Profile{})
	}
	profiles[len(profiles)-1].FirstName = parts[0]
	if len(parts) > 1 {
		profiles[len(profiles)-1].LastName = parts[1]
	}
	return profiles
}

func AddFirstName(profiles []*Profile, v string) []*Profile {
	if len(profiles) == 0 || profiles[len(profiles)-1].FirstName != "" {
		profiles = append(profiles, &Profile{})
	}
	profiles[len(profiles)-1].FirstName = v
	return profiles
}

func AddLastName(profiles []*Profile, v string) []*Profile {
	if len(profiles) == 0 || profiles[len(profiles)-1].LastName != "" {
		profiles = append(profiles, &Profile{})
	}
	profiles[len(profiles)-1].LastName = v
	return profiles
}

func AddUsername(profiles []*Profile, v string) []*Profile {
	if len(profiles) == 0 || profiles[len(profiles)-1].Username != "" {
		profiles = append(profiles, &Profile{})
	}
	profiles[len(profiles)-1].Username = v
	return profiles
}

func AddGender(profiles []*Profile, v string) []*Profile {
	if len(profiles) == 0 || profiles[len(profiles)-1].Gender != "" {
		profiles = append(profiles, &Profile{})
	}
	profiles[len(profiles)-1].Gender = v
	return profiles
}
