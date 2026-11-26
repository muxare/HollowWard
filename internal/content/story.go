package content

// OpeningLetter is the previous warden's last letter, shown when a tenure
// begins.
var OpeningLetter = []string{
	"To whoever holds the keys after me.",
	"The city is a lid. I was told this on my first day and did not believe it either.",
	"You have three stewards and six districts. That is not an oversight. It is the job.",
	"What you neglect by day, it takes at dusk. What it takes, it remakes out of what was missing.",
	"Send people in. They will come back tired, or marked, or not at all. Send them anyway.",
	"I kept a Ledger of what grows from what. It is yours now. It is the only thing I can leave you.",
	"I can hear it turning over. I am going down to look.",
}

// Fragments are the previous warden's notes, found one at a time in safe
// rooms. Finding all five unlocks ClosingText.
var Fragments = []string{
	"Third night. A hollow is not a ruin. It is the district, dreamed by something that only knows what the district lacked.",
	"The factor was in there. I knew his coat. He had not been paid in bread for a month, and that was all he was any more.",
	"I let the Mills starve for a week, on purpose, and harvested what came. It worked. I do not like how well it worked.",
	"The rooms we kept in repair stay kind, even down there. A swept floor is a sort of prayer.",
	"It is not angry. I want that written down. It is asleep, and we are the thing it is lying on.",
}

// ClosingText is unlocked when all five fragments have been found.
var ClosingText = []string{
	"The last page is in a different hand. Yours, perhaps, later.",
	"I went down. There is a room at the bottom with nothing missing from it. Every need at its mark, every wall plumb.",
	"It dreams that room too. It dreams it the way you dream of home.",
	"I do not think it wants to wake. I think it wants someone to keep the city so well that it never has to.",
	"I could not. Twenty nights is a long time. But the Ledger is thicker now. Keep it. Add to it. Hand it on.",
}

// DreamTiers are the Dreamer's nightly lines, grouped by how high the meter
// stands. DreamTierAt in tuning.go says where each group begins.
var DreamTiers = [][]string{
	{
		"It sleeps. You can tell by the quiet.",
		"Under the flagstones, something turns over.",
		"A long breath from below. Then nothing.",
		"The wells are still tonight.",
	},
	{
		"It dreams of a city. This one, nearly.",
		"Somewhere below, a door you do not have is opened.",
		"The dogs will not cross the market square.",
		"You hear your own name, said the way a stranger would say it.",
	},
	{
		"It dreams of you now. You are smaller there.",
		"The lamps lean toward the ground.",
		"Every bell in the city rings a little flat.",
		"It has started counting your stewards.",
	},
	{
		"It knows which districts you let go.",
		"The streets are warm underfoot.",
		"Its breath comes up through the drains.",
		"It dreams the walls thinner.",
	},
	{
		"It is nearly awake. It is being polite about it.",
		"The lid creaks. Everyone agrees it is the wind.",
		"It has stopped dreaming of the city. It is dreaming of the sky.",
		"There is light under the cobbles.",
	},
	{
		"It opens one eye.",
		"It says: enough.",
	},
}

// DreamLine is the Dreamer's line for a night, given the meter in points.
func DreamLine(points, night int) string {
	tier := 0
	for i, at := range T.DreamTierAt {
		if points >= at {
			tier = i
		}
	}
	lines := DreamTiers[min(tier, len(DreamTiers)-1)]
	return lines[night%len(lines)]
}
