package utils

import (
	"strings"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

func TypePullRequest(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))

	switch {
	case strings.HasPrefix(title, domain.TypeFeat):
		return domain.TypeFeat
	case strings.HasPrefix(title, domain.TypeFix):
		return domain.TypeFix
	case strings.HasPrefix(title, domain.TypeHot):
		return domain.TypeHot
	case strings.HasPrefix(title, domain.TypeDoc):
		return domain.TypeDoc
	case strings.HasPrefix(title, domain.TypeChore):
		return domain.TypeChore
	default:
		return domain.TypeOther
	}
}
