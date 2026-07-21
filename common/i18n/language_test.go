package i18n

import (
	"testing"
)

func TestTranslateMessage(t *testing.T) {
	SetLanguage("en")
	if got := T("start http on "); got != "start http on " {
		t.Fatalf("expected english fallback, got %q", got)
	}

	SetLanguage("zh")
	if got := T("start http on "); got != "启动 HTTP 服务于 " {
		t.Fatalf("expected chinese translation, got %q", got)
	}
}
