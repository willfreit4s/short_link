package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewShortLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		original   string
		expected   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "adds https scheme to naked url",
			original: "example.com",
			expected: "https://example.com",
		},
		{
			name:     "keeps https scheme",
			original: "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "keeps http scheme",
			original: "http://example.com",
			expected: "http://example.com",
		},
		{
			name:       "rejects empty url",
			original:   "   ",
			wantErr:    true,
			wantErrMsg: ErrOriginalURLIsRequired.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			shortLink, err := NewShortLink(tt.original)
			if tt.wantErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, shortLink.ID)
			require.Equal(t, tt.expected, shortLink.OriginalURL)
			require.False(t, shortLink.CreatedAt.IsZero())
		})
	}
}

func TestShortLinkValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid link with generated id", func(t *testing.T) {
		t.Parallel()

		link := &ShortLink{
			ID:          "ABCDEFGHIJ",
			OriginalURL: "https://example.com",
		}

		require.NoError(t, link.Validate())
	})

	t.Run("rejects missing id", func(t *testing.T) {
		t.Parallel()

		link := &ShortLink{OriginalURL: "https://example.com"}
		require.EqualError(t, link.Validate(), ErrIDIsRequired.Error())
	})

	t.Run("rejects empty original url", func(t *testing.T) {
		t.Parallel()

		link := &ShortLink{ID: "ABCDEFGHIJ"}
		require.EqualError(t, link.Validate(), ErrOriginalURLIsRequired.Error())
	})
}
