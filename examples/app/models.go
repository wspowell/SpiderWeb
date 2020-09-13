package app

type MyRequestBodyModel struct {
	MyString   string `json:"my_string"`
	MyInt      int    `json:"my_int"`
	ShouldFail bool   `json:"fail"`
}

type MyResponseBodyModel struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
}
