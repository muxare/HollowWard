package sim

import (
	"fmt"
	"strings"

	"hollowward/internal/content"
)

// Chronicle tells a tenure back as a short history, one entry per night worth
// telling. It reads the event log and nothing else.
func Chronicle(events []Event) []string {
	var out []string
	var line []string // sentences about the current night
	night, quietFrom := 0, 0
	var party []string

	flushQuiet := func(upTo int) {
		switch quietFrom {
		case 0:
		case upTo:
			out = append(out, fmt.Sprintf("Night %d was quiet.", upTo))
		default:
			out = append(out, fmt.Sprintf("Nights %d to %d were quiet.", quietFrom, upTo))
		}
		quietFrom = 0
	}
	flush := func() {
		if len(line) == 0 {
			if night > 0 && quietFrom == 0 {
				quietFrom = night
			}
			return
		}
		flushQuiet(night - 1)
		out = append(out, fmt.Sprintf("Night %d. %s", night, strings.Join(line, " ")))
		line = nil
	}
	the := func(d content.DistrictID) string { return "the " + content.Districts[d].Name }
	names := func() string {
		switch len(party) {
		case 0:
			return "Nobody"
		case 1:
			return party[0]
		}
		return strings.Join(party[:len(party)-1], ", ") + " and " + party[len(party)-1]
	}

	for _, e := range events {
		switch e.Kind {
		case EvTenureStarted:
			out = append(out, "A new warden took the keys.")
		case EvDayStarted:
			flush()
			night, party = e.Day, nil
		case EvRuined:
			line = append(line, fmt.Sprintf("%s in %s fell into ruin.", capital(thingName(e.District, e.ID)), the(e.District)))
		case EvUpgradeBought:
			line = append(line, fmt.Sprintf("The city made %s.", thingName(e.District, e.ID)))
		case EvTurned:
			why := ""
			if len(e.Causes) > 0 {
				why = ": " + e.Causes[0].Describe()
			}
			line = append(line, fmt.Sprintf("%s turned, at Neglect %d%s.", capital(the(e.District)), e.N, why))
		case EvJoined:
			party = append(party, e.Who)
		case EvFragment:
			line = append(line, "In a room that had been kept, a page in the old warden's hand.")
		case EvHeartCleared:
			line = append(line, fmt.Sprintf("%s went in and put down %s; %s %s a district again.", names(), thingName(e.District, e.ID), the(e.District), e.District.Verb("was", "were")))
		case EvWentHome:
			line = append(line, fmt.Sprintf("%s went into %s and came home with its heart still beating.", names(), the(e.District)))
		case EvWiped:
			line = append(line, fmt.Sprintf("%s went into %s. None of them walked out.", names(), the(e.District)))
		case EvDrivenOut:
			line = append(line, fmt.Sprintf("%s went into %s and were driven out.", names(), the(e.District)))
		case EvDied:
			line = append(line, fmt.Sprintf("%s did not wake.", e.Who))
		case EvScarred:
			line = append(line, fmt.Sprintf("%s came back marked: %s.", e.Who, thingName(e.District, e.ID)))
		case EvTenureEnded:
			if Outcome(e.ID) == LostLid && len(line) == 0 {
				flushQuiet(night - 1) // the lid broke at dusk: tonight never happened
			} else {
				flush()
				flushQuiet(night)
			}
			switch Outcome(e.ID) {
			case Won:
				out = append(out, fmt.Sprintf("%d nights. The lid held. The keys pass on.", e.N))
			case LostLid:
				out = append(out, fmt.Sprintf("On night %d the lid broke. The warden had kept it %d nights.", e.N+1, e.N))
			case LostNoParty:
				out = append(out, fmt.Sprintf("After night %d there was no one left to send.", e.N))
			}
			night = 0
		}
	}
	flush()
	return out
}

func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
