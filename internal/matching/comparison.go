package matching

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

func compare(input Input, recording *Recording) (Candidate, bool) {
	const (
		nameWeight     = 0.45
		albumWeight    = 0.05
		durationWeight = 0.05
	)

	titleMatches := normalize(input.Title) == normalize(recording.Title)
	artistMatches := creditMatches(input.Artist, recording.ArtistCredit)
	candidate := Candidate{Recording: *recording}
	if titleMatches {
		candidate.Score += nameWeight
	}
	if artistMatches {
		candidate.Score += nameWeight
	}
	if albumMatches(input.Album, recording.Albums) {
		candidate.Score += albumWeight
	}

	duration := compareDuration(input.Duration, recording.Duration)
	if duration.known && duration.matches {
		candidate.Score += durationWeight
	}

	return candidate, titleMatches && artistMatches && (!duration.known || duration.matches)
}

type durationEvidence struct {
	known   bool
	matches bool
}

func compareDuration(input, recording time.Duration) durationEvidence {
	const (
		durationTolerance = 3 * time.Second
		unknown           = 0
	)

	known := input > unknown && recording > unknown
	difference := input - recording

	return durationEvidence{known: known, matches: difference >= -durationTolerance && difference <= durationTolerance}
}

func creditMatches(artist string, credits []ArtistCredit) bool {
	var creditedNames, canonicalNames []string
	for _, credit := range credits {
		creditedNames = append(creditedNames, credit.Name, credit.JoinPhrase)
		canonicalNames = append(canonicalNames, credit.Artists.Name, credit.JoinPhrase)
	}

	input := normalize(artist)

	return input == normalize(strings.Join(creditedNames, "")) ||
		input == normalize(strings.Join(canonicalNames, ""))
}

func albumMatches(album string, albums []Album) bool {
	if album == "" {
		return false
	}

	for _, release := range albums {
		if normalize(album) == normalize(release.Title) {
			return true
		}
	}

	return false
}

func normalize(value string) string {
	return strings.Join(strings.Fields(cases.Fold().String(norm.NFC.String(value))), " ")
}
