package dto

type GetShortLinkInput struct {
	Hash string
}

type GetShortLinkOutput struct {
	OriginalURL string
}
