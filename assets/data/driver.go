package data

type (
	Driver interface {
		IsForeignKeyViolation(err error) bool
		IsUniqueViolation(err error) bool
	}
)
