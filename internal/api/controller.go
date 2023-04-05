package api

const (
	MIMEApplicationJSON = "application/json"
)

type Controller struct {
}

func NewController() Controller {
	return Controller{}
}

type Employee struct {
	Id   int
	Name string
}
type Response struct {
	Error    string
	Employee Employee
}
