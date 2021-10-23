package test


type UserModel struct {
	Id uint
	Name string
}

type User struct {
	changes map[string]interface{}
	UserModel //db
}
