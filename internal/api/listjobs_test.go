package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListJobs_QueryValidation(t *testing.T) {
	t.Parallel()
	h := &Handler{}
	cases := []struct {
		name       string
		rawQuery   string
		wantStatus int
	}{
		{
			name:       "missing status",
			rawQuery:   "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported status",
			rawQuery:   "status=pending",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid limit",
			rawQuery:   "status=failed&limit=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "negative offset",
			rawQuery:   "status=failed&offset=-1",
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/jobs?"+tc.rawQuery, nil)
			w := httptest.NewRecorder()
			h.ListJobs(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("status %d, body %s", w.Code, w.Body.String())
			}
		})
	}
}
