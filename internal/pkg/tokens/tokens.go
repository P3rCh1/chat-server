package tokens

type TokenProvider interface {
	Gen(id int) (string, error)
	Verify(tokenString string) (int, error)
}
