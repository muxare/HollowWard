package sim

import (
	"fmt"
	"strings"

	"hollowward/internal/content"
)

// thingName turns a slug from the event log back into a name for people.
func thingName(d content.DistrictID, slug string) string {
	for _, e := range content.Enemies {
		if e.Slug == slug {
			return e.Name
		}
	}
	if d.Valid() {
		for n := content.Need(0); n < content.NumNeeds; n++ {
			if slug == content.BossSlug(d, n) {
				return content.BossName(d, n)
			}
		}
		for _, b := range content.Districts[d].Buildings {
			if b.Slug == slug {
				return "the " + b.Name
			}
		}
	}
	for _, c := range content.EventCards {
		if c.Slug == slug {
			return c.Title
		}
	}
	for _, dec := range content.Decrees {
		if dec.Slug == slug {
			return dec.Name
		}
	}
	for _, a := range content.Abilities {
		if a.Slug == slug {
			return a.Name
		}
	}
	for _, u := range content.Upgrades {
		if u.Slug == slug {
			return u.Name
		}
	}
	for _, sc := range content.Scars {
		if sc.Slug == slug {
			return sc.Name
		}
	}
	return slug
}

// plural names n enemies of a kind: "4 Gaunts", "a Gaunt".
func plural(slug string, n int, d content.DistrictID) string {
	for _, e := range content.Enemies {
		if e.Slug == slug {
			if n == 1 {
				return "a " + e.Name
			}
			return fmt.Sprintf("%d %s", n, e.Plural)
		}
	}
	return thingName(d, slug)
}

func because(causes []Cause) string {
	if len(causes) == 0 {
		return ""
	}
	parts := make([]string, 0, len(causes))
	for _, c := range causes {
		parts = append(parts, c.Describe())
	}
	return " (" + strings.Join(parts, "; ") + ")"
}

// Text renders an event as one plain line. Every line that names something
// generated also says why it exists.
func (e Event) Text() string {
	where := ""
	if e.District.Valid() {
		where = "the " + content.Districts[e.District].Name
	}
	switch e.Kind {
	case EvTenureStarted:
		return "A new warden takes the keys."
	case EvDayStarted:
		return fmt.Sprintf("--- Day %d ---", e.Day)
	case EvCardDrawn:
		return fmt.Sprintf("Event: %s.", thingName(e.District, e.ID))
	case EvCardResolved:
		for _, c := range content.EventCards {
			if c.Slug == e.ID && e.N < len(c.Choices) {
				return fmt.Sprintf("You chose: %s. %s", c.Choices[e.N].Label, c.Choices[e.N].Result)
			}
		}
	case EvDecreeIssued:
		if where != "" {
			return fmt.Sprintf("Decree: %s, for %s.", thingName(e.District, e.ID), where)
		}
		return fmt.Sprintf("Decree: %s.", thingName(e.District, e.ID))
	case EvStewardWorked:
		return fmt.Sprintf("%s worked in %s: %s +%d.", content.StewardNames[e.V], where, strings.ToUpper(e.ID[:1])+e.ID[1:], e.N)
	case EvRepaired:
		return fmt.Sprintf("%s repaired %s in %s: +%d.", content.StewardNames[e.V], thingName(e.District, e.ID), where, e.N)
	case EvProduced:
		return fmt.Sprintf("%s produced %d %s.", strings.ToUpper(where[:1])+where[1:], e.N, e.ID)
	case EvRuined:
		return fmt.Sprintf("%s in %s fell into ruin (condition %d).", strings.ToUpper(thingName(e.District, e.ID)[:1])+thingName(e.District, e.ID)[1:], where, e.N)
	case EvDayEnded:
		return "Dusk."
	case EvTurned:
		return fmt.Sprintf("%s TURNS. Neglect %d%s.", strings.ToUpper(where), e.N, because(e.Causes))
	case EvQuietNight:
		return "No district turns tonight."
	case EvPartySent:
		return fmt.Sprintf("A party of %d enters the hollow of %s.", e.N, where)
	case EvJoined:
		return fmt.Sprintf("  %s goes in.", e.Who)
	case EvStayedHome:
		return "Nobody goes in tonight."
	case EvRoomEntered:
		return fmt.Sprintf("Room: %s%s.", e.ID, because(e.Causes))
	case EvMet:
		if e.V == 1 {
			return fmt.Sprintf("  Waiting here: %s%s.", thingName(e.District, e.ID), because(e.Causes))
		}
		return fmt.Sprintf("  Waiting here: %s%s.", plural(e.ID, e.N, e.District), because(e.Causes))
	case EvHazard:
		return fmt.Sprintf("  Hazard: everyone loses %d HP%s.", e.N, because(e.Causes))
	case EvHaven:
		return fmt.Sprintf("  Safe room: the party rests and recovers %d HP%s.", e.N, because(e.Causes))
	case EvFightStarted:
		return "  A fight."
	case EvRound:
		return fmt.Sprintf("  Round %d.", e.N)
	case EvCitizenActed:
		switch e.ID {
		case "defend":
			return fmt.Sprintf("    %s braces.", e.Who)
		case "attack":
			return fmt.Sprintf("    %s strikes for %d.", e.Who, e.V)
		}
		if e.V > 0 {
			return fmt.Sprintf("    %s uses %s for %d.", e.Who, thingName(e.District, e.ID), e.V)
		}
		return fmt.Sprintf("    %s uses %s.", e.Who, thingName(e.District, e.ID))
	case EvEnemyActed:
		switch {
		case e.ID == "stunned":
			return "    An enemy is stunned and loses its turn."
		case e.Who != "":
			return fmt.Sprintf("    Enemy %s: %s takes %d.", e.ID, e.Who, e.V)
		case e.V > 0:
			return fmt.Sprintf("    Enemy %s: %d damage in all.", e.ID, e.V)
		}
		return fmt.Sprintf("    Enemy %s.", e.ID)
	case EvEnemyFell:
		return fmt.Sprintf("    %s falls.", thingName(e.District, e.ID))
	case EvDowned:
		return fmt.Sprintf("    %s is DOWNED.", e.Who)
	case EvRoomCleared:
		return fmt.Sprintf("  %s is cleared.", e.ID)
	case EvLoot:
		if e.Who != "" {
			return fmt.Sprintf("  %s scavenges %d %s.", e.Who, e.N, e.ID)
		}
		return fmt.Sprintf("  Loot: %d %s.", e.N, e.ID)
	case EvHeartCleared:
		return fmt.Sprintf("%s is dead. The heart of %s is still.", thingName(e.District, e.ID), where)
	case EvRestored:
		return fmt.Sprintf("%s is a district again.", strings.ToUpper(where[:1])+where[1:])
	case EvWentHome:
		return fmt.Sprintf("The party goes home. %s stays turned.", strings.ToUpper(where[:1])+where[1:])
	case EvWiped:
		return "The whole party is down. They are carried out."
	case EvDrivenOut:
		return "The fight drags on too long. The hollow drives the party out."
	case EvDied:
		return fmt.Sprintf("%s does not wake.", e.Who)
	case EvScarred:
		return fmt.Sprintf("%s lives, marked: %s.", e.Who, thingName(e.District, e.ID))
	case EvExhausted:
		return fmt.Sprintf("%s will work at half strength tomorrow.", e.Who)
	case EvNightEnded:
		return "Dawn."
	case EvDreamerRose:
		return fmt.Sprintf("The Dreamer rises by %d.%d, to %d.%d of %d.", e.N/10, e.N%10, e.V/10, e.V%10, content.T.DreamerMax/10)
	case EvDreamerSoothed:
		return fmt.Sprintf("The Dreamer settles by %d.%d, to %d.%d.", e.N/10, e.N%10, e.V/10, e.V%10)
	case EvDreamLine:
		return "  \"" + e.ID + "\""
	case EvUpgradeBought:
		if e.District.Valid() {
			return fmt.Sprintf("The city makes a %s for %s in %s.", thingName(e.District, e.ID), "the "+content.Districts[e.District].Buildings[e.V].Name, where)
		}
		return fmt.Sprintf("The city makes %s.", thingName(e.District, e.ID))
	case EvFragment:
		return fmt.Sprintf("  In %s, a page in the old warden's hand: \"%s\"", e.ID, content.Fragments[e.N%len(content.Fragments)])
	case EvTenureEnded:
		switch Outcome(e.ID) {
		case Won:
			return fmt.Sprintf("THE TENURE ENDS. %d nights survived. The lid held.", e.N)
		case LostLid:
			return fmt.Sprintf("THE LID BREAKS. The tenure ends after %d nights.", e.N)
		case LostNoParty:
			return fmt.Sprintf("NO ONE IS LEFT TO SEND. The tenure ends after %d nights.", e.N)
		}
	}
	return ""
}
