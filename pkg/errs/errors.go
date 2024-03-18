package errs

type ErrorResponse struct {
	Errors []ErrorDetail `json:"errors"`
	Status int
}

type ErrorDetail struct {
	Param string `json:"param"`
	Msg   string `json:"msg"`
}

const (
	NotFound            = "not found"
	NoAccess            = "no access"
	ShortPass           = "short password"
	UserExistError      = "user already exists"
	BadPass             = "invalid password"
	BadID               = "bad ID"
	JSONerror           = "decode JSON error"
	SessionError        = "error session creation"
	UnauthorizedError   = "user unauthorized"
	EmptyActorError     = "empty actor"
	DatabaseError       = "DB error"
	FailPassing         = "failed to parse token"
	InvalidToken        = "invalid token"
	UserClaimsError     = "user not found in token claims"
	UserNotExist        = "user not exist"
	ReadingOrderError   = "error reading order"
	ReadingOrderByError = "incorrect orderBy"
	EmptySearchError    = "empty search"
	WrongColumnError    = "wrong column name"
	EmptyUsernameError  = "empty username"
	HashPasswordError   = "failed to hash password"
)
