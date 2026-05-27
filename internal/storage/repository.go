package storage

type Repository interface {
	Save() error
	GetByShort() (string, error)
	GetByOriginal() (string, error)
}
