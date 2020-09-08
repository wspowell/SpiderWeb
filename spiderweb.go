package spiderweb

import (
	_ "spiderweb/endpoint"
)

type framework struct {
}

func New() *framework {
	return &framework{}
}
