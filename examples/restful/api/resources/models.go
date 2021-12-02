package resources

type MyRequestBodyModel struct {
	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type MyResponseBodyModel struct {
	MyString string `json:"outputString"`
	MyInt    int    `json:"outputInt"`
}
