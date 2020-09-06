package spiderweb

import (
	_ "spiderweb/endpoint_tags"
)

type framework struct {
}

func New() *framework {
	return &framework{}
}
