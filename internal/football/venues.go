package football

// venueInfo holds stadium metadata derived from the home team name.
// football-data.org free tier does not return a venue field, so we map
// home team name → stadium, city, country.
type venueInfo struct {
	Stadium string
	City    string
	Country string
}

// venues maps football-data.org homeTeam.name to stadium metadata.
// Covers Premier League clubs, Champions League regulars and Liverpool FC itself.
var venues = map[string]venueInfo{
	// Liverpool
	"Liverpool FC": {"Anfield", "Liverpool", "England"},

	// Premier League
	"Arsenal FC":                      {"Emirates Stadium", "London", "England"},
	"Aston Villa FC":                  {"Villa Park", "Birmingham", "England"},
	"AFC Bournemouth":                 {"Vitality Stadium", "Bournemouth", "England"},
	"Brentford FC":                    {"Gtech Community Stadium", "London", "England"},
	"Brighton & Hove Albion FC":       {"Amex Stadium", "Brighton", "England"},
	"Chelsea FC":                      {"Stamford Bridge", "London", "England"},
	"Crystal Palace FC":               {"Selhurst Park", "London", "England"},
	"Everton FC":                       {"Goodison Park", "Liverpool", "England"},
	"Fulham FC":                        {"Craven Cottage", "London", "England"},
	"Ipswich Town FC":                 {"Portman Road", "Ipswich", "England"},
	"Leicester City FC":               {"King Power Stadium", "Leicester", "England"},
	"Manchester City FC":              {"Etihad Stadium", "Manchester", "England"},
	"Manchester United FC":            {"Old Trafford", "Manchester", "England"},
	"Newcastle United FC":             {"St. James' Park", "Newcastle", "England"},
	"Nottingham Forest FC":            {"The City Ground", "Nottingham", "England"},
	"Southampton FC":                  {"St Mary's Stadium", "Southampton", "England"},
	"Tottenham Hotspur FC":            {"Tottenham Hotspur Stadium", "London", "England"},
	"West Ham United FC":              {"London Stadium", "London", "England"},
	"Wolverhampton Wanderers FC":      {"Molineux Stadium", "Wolverhampton", "England"},

	// Champions League regulars
	"Real Madrid CF":          {"Santiago Bernabéu", "Madrid", "Spain"},
	"FC Barcelona":            {"Estadi Olímpic Lluís Companys", "Barcelona", "Spain"},
	"Club Atlético de Madrid": {"Cívitas Metropolitano", "Madrid", "Spain"},
	"FC Bayern München":       {"Allianz Arena", "Munich", "Germany"},
	"Borussia Dortmund":       {"Signal Iduna Park", "Dortmund", "Germany"},
	"Paris Saint-Germain FC":  {"Parc des Princes", "Paris", "France"},
	"Juventus FC":             {"Allianz Stadium", "Turin", "Italy"},
	"AC Milan":                {"San Siro", "Milan", "Italy"},
	"FC Internazionale Milano": {"San Siro", "Milan", "Italy"},
	"SSC Napoli":              {"Stadio Diego Armando Maradona", "Naples", "Italy"},
	"AS Roma":                 {"Stadio Olimpico", "Rome", "Italy"},
	"SL Benfica":              {"Estádio da Luz", "Lisbon", "Portugal"},
	"FC Porto":                {"Estádio do Dragão", "Porto", "Portugal"},
	"Ajax":                    {"Johan Cruyff Arena", "Amsterdam", "Netherlands"},
	"Celtic FC":               {"Celtic Park", "Glasgow", "Scotland"},
	"Rangers FC":              {"Ibrox Stadium", "Glasgow", "Scotland"},
}

// lookupVenue returns stadium metadata for the given home team name.
// Returns zero-value venueInfo if the team is not found.
func lookupVenue(homeTeamName string) venueInfo {
	return venues[homeTeamName]
}
