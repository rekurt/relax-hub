package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectImportFormat(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		want        string
	}{
		{"xlsx by extension", "listings.xlsx", "", "xlsx"},
		{"csv by extension", "listings.csv", "", "csv"},
		{"xlsx by content type", "file", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "xlsx"},
		{"csv by content type", "file", "text/csv", "csv"},
		{"csv by text/plain", "file", "text/plain", "csv"},
		{"extension takes priority", "data.xlsx", "text/csv", "xlsx"},
		{"unknown format", "file.txt", "application/octet-stream", ""},
		{"uppercase extension", "DATA.XLSX", "", "xlsx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectImportFormat(tt.filename, tt.contentType)
			assert.Equal(t, tt.want, got)
		})
	}
}
