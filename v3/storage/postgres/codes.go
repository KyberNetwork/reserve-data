package postgres

//https://www.postgresql.org/docs/9.5/errcodes-appendix.html
const (
	errCodeUniqueViolation = "23505"
	errCodeCheckViolation  = "23514"
	errForeignKeyViolation = "23503"
	errAssertFailure       = "P0004"
	// custom error
	errQuoteInvalid = "KEQUO"
	errBaseInvalid  = "KEBAS"
)
