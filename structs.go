package main

type Arbitraje struct {
	Fallos []*Fallo `json:"fallos"`
}

type Fallo struct {
	NombreDominio    string   `json:"nombreDominio"`
	LstPartes        []*Parte `json:"lstPartes"`
	GanaRevocante    int      `json:"ganaRevocante"`
	ArchivoSentencia string   `json:"archivoSentencia"`
}

type Parte struct {
	Nombre string `json:"nombre"`
}

func (f *Fallo) AdjustNames() {
	numParts := len(f.LstPartes)
	availableCharsPerDomain := int((MAX_SENTENCE_NAMES_LENGTH - (4 * numParts) - len(f.NombreDominio)) / numParts)
	for _, part := range f.LstPartes {
		if len(part.Nombre) > availableCharsPerDomain {
			part.Nombre = part.Nombre[:availableCharsPerDomain-1] + "…"
		}
	}
}
