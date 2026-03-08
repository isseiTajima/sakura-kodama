package llm

import "testing"

func TestFallbackSpeech_UserClickIsUpdatedPhrase(t *testing.T) {
	SetSeed(42)
	speech := FallbackSpeech(ReasonUserClick)

	if speech == "" || speech == "…" {
		t.Errorf("want specific text for user click, got %q", speech)
	}
}

func TestFallbackSpeech_UnknownReasonUsesEllipsis(t *testing.T) {
	speech := FallbackSpeech(Reason("unknown"))

	if speech != "…" {
		t.Errorf("want ellipsis fallback for unknown reason, got %q", speech)
	}
}

func TestFallbackSpeech_GitCommitIsNonEmpty(t *testing.T) {
	speech := FallbackSpeech(ReasonGitCommit)
	if speech == "" || speech == "…" {
		t.Errorf("want specific text for %s, got %q", ReasonGitCommit, speech)
	}
}
