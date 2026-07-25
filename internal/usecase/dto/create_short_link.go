package dto

type CreateShortLinkInput struct {
	OriginalURL string
}

type CreateShortLinkOutput struct {
	ID          string
	Hash        string
	OriginalURL string
}
