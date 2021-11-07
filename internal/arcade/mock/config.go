package mock

// Logger is a mock of the Logger config.
type Logger struct {
	Level_, File_, Format_ string
}

// Level returns the mocked logging level.
func (m Logger) Level() string {
	return m.Level_
}

// File returns the mocked logging output file.
func (m Logger) File() string {
	return m.File_
}

// Format returns the mocked logging format.
func (m Logger) Format() string {
	return m.Format_
}

// DB is a mock of the db config.
type DB struct {
	DSN_, DriverName_ string
}

// DriverName returns the mocked db driver name.
func (d DB) DriverName() string {
	return d.DriverName_
}

// DSN returns the mocked db DSN.
func (d DB) DSN() string {
	return d.DSN_
}

// Server is a mock of the Server config.
type Server struct {
	Addr_, Cert_, Key_, CACert_ string
}

// Addr returns the mocked server address.
func (s Server) Addr() string {
	return s.Addr_
}

// Cert returns the mocked server certificate.
func (s Server) Cert() string {
	return s.Cert_
}

// Key returns the mocked server certificate private key.
func (s Server) Key() string {
	return s.Key_
}

// CACert returns the mocked server CA cert.
func (s Server) CACert() string {
	return s.CACert_
}
