package utils

import (
	"strings"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

func TypePullRequest(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))

	switch {
	case strings.HasPrefix(title, "feat"):
		return domain.TypeFeat
	case strings.HasPrefix(title, "fix"):
		return domain.TypeFix
	case strings.HasPrefix(title, "hot"):
		return domain.TypeHot
	case strings.HasPrefix(title, "doc"):
		return domain.TypeDoc
	case strings.HasPrefix(title, "chore"):
		return domain.TypeChore
	default:
		return domain.TypeOther
	}
}
