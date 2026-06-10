package football

// Translations mapeia o nome em inglês retornado pela API para português.
var Translations = map[string]string{
	"Uruguay":            "Uruguai",
	"Germany":            "Alemanha",
	"Spain":              "Espanha",
	"Paraguay":           "Paraguai",
	"Argentina":          "Argentina",
	"Ghana":              "Gana",
	"Brazil":             "Brasil",
	"Portugal":           "Portugal",
	"Japan":              "Japão",
	"Mexico":             "México",
	"England":            "Inglaterra",
	"United States":      "Estados Unidos",
	"South Korea":        "Coreia do Sul",
	"France":             "França",
	"South Africa":       "África do Sul",
	"Algeria":            "Argélia",
	"Australia":          "Austrália",
	"New Zealand":        "Nova Zelândia",
	"Switzerland":        "Suíça",
	"Ecuador":            "Equador",
	"Sweden":             "Suécia",
	"Czechia":            "República Tcheca",
	"Croatia":            "Croácia",
	"Saudi Arabia":       "Arábia Saudita",
	"Tunisia":            "Tunísia",
	"Turkey":             "Turquia",
	"Senegal":            "Senegal",
	"Belgium":            "Bélgica",
	"Morocco":            "Marrocos",
	"Austria":            "Áustria",
	"Colombia":           "Colômbia",
	"Egypt":              "Egito",
	"Canada":             "Canadá",
	"Haiti":              "Haiti",
	"Iran":               "Irã",
	"Bosnia-Herzegovina": "Bósnia e Herzegovina",
	"Panama":             "Panamá",
	"Cape Verde Islands": "Cabo Verde",
	"Congo DR":           "Rep. Democrática do Congo",
	"Ivory Coast":        "Costa do Marfim",
	"Qatar":              "Catar",
	"Jordan":             "Jordânia",
	"Iraq":               "Iraque",
	"Uzbekistan":         "Uzbequistão",
	"Netherlands":        "Holanda",
	"Norway":             "Noruega",
	"Scotland":           "Escócia",
	"Curaçao":            "Curaçao",
}

// Translate retorna o nome em português ou o original se não tiver tradução
func Translate(name string) string {
	if pt, ok := Translations[name]; ok {
		return pt
	}
	return name
}
