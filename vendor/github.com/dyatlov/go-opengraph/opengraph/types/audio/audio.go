package audio

// Audio defines Open Graph Audio Type
type Audio struct {
	URL       string `json:"url"`
	SecureURL string `json:"secure_url"`
	Type      string `json:"type"`
}

func NewAudio() *Audio {
	return &Audio{}
}

func AddUrl(audios []*Audio, v string) []*Audio {
	if len(audios) == 0 || audios[len(audios)-1].URL != "" {
		audios = append(audios, &Audio{})
	}
	audios[len(audios)-1].URL = v
	return audios
}

func AddSecureUrl(audios []*Audio, v string) []*Audio {
	if len(audios) == 0 || audios[len(audios)-1].SecureURL != "" {
		audios = append(audios, &Audio{})
	}
	audios[len(audios)-1].SecureURL = v
	return audios
}

func AddType(audios []*Audio, v string) []*Audio {
	if len(audios) == 0 || audios[len(audios)-1].Type != "" {
		audios = append(audios, &Audio{})
	}
	audios[len(audios)-1].Type = v
	return audios
}
