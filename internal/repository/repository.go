package repository

type Repository interface {
	GetOrCreate(token, url string) (string, bool, error)
	GetURL(token string) (string, error)
	GetToken(url string) (string, error)
}
