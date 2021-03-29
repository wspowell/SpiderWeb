package app

import (
	"io"
	"net/http"
	"time"

	"github.com/wspowell/spiderweb"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/examples/auth"
	"github.com/wspowell/spiderweb/examples/error_handlers"
	"github.com/wspowell/spiderweb/examples/validators"
	"github.com/wspowell/spiderweb/logging"
)

type NoopLogConfig struct {
	*logging.Config
}

func (self *NoopLogConfig) Out() io.Writer {
	return io.Discard
}

func SetupServer() spiderweb.Server {
	serverConfig := spiderweb.NewServerConfig("localhost", 8080, endpoint.Config{
		Auther:            auth.Noop{},
		ErrorHandler:      error_handlers.ErrorJsonWithCodeResponse{},
		LogConfig:         &NoopLogConfig{logging.NewConfig(logging.LevelDebug, map[string]interface{}{})},
		MimeTypeHandlers:  map[string]endpoint.MimeTypeHandler{},
		RequestValidator:  validators.NoopRequest{},
		ResponseValidator: validators.NoopResponse{},
		Timeout:           30 * time.Second,
	})

	serverConfig.Handle(http.MethodPost, "/resources", &PostResource{})
	serverConfig.Handle(http.MethodGet, "/resources/{id}", &GetResource{})

	return spiderweb.NewServer(serverConfig)
}

func swaggerSpec() []byte {
	return []byte(`
	{
		"swagger":"2.0",
		"info":{
			"version":"1.0.0",
			"title":"Example App",
			"contact":{
				"name":"Wesley Powell",
				"url":"http://github.com/wspowell/spiderweb/examples/app"
			}
		},
		"host":"example.com",
		"basePath":"/",
		"schemes":[
			"http"
		],
		"securityDefinitions":{
			"basic":{
				"type":"basic"
			}
		},
		"paths":{
			"/resources":{
				"post":{
					"security":[
						{
							"basic":[]
						}
					],
					"tags":[
						"Resources"
					],
					"operationId":"createResource",
					"summary":"Creates a new resource",
					"consumes":[
						"application/json"
					],
					"produces":[
						"application/json"
					],
					"parameters":[
						{
							"name":"resource",
							"in":"body",
							"description":"The resource to create",
							"required":true,
							"schema":{
								"$ref":"#/definitions/CreateResourceRequest"
							}
						}
					],
					"responses":{
						"200":{
							"description":"Created Pet response",
							"schema":{
								"$ref":"#/definitions/ResourceResponse"
							}
						},
						"500":{
							"description":"Internal server error",
							"schema":{
								"$ref":"#/definitions/ErrorResponse"
							}
						}
					}
				}
			},
			"/resources/{id}":{
				"get":{
					"security":[
						{
							"basic":[]
						}
					],
					"tags":[
						"Resources"
					],
					"operationId":"getResource",
					"summary":"Receives a resource by ID",
					"produces":[
						"application/json"
					],
					"parameters":[
						{
							"name":"id",
							"in":"path",
							"required":true,
							"description":"The resource ID",
							"type":"integer",
							"format":"int64"
						}
					],
					"responses":{
						"200":{
							"description":"Resource Response",
							"schema":{
								"$ref":"#/definitions/ResourceResponse"
							}
						},
						"500":{
							"description":"Internal server error",
							"schema":{
								"$ref":"#/definitions/ErrorResponse"
							}
						}
					}
				}
			}
		},
		"definitions":{
			"CreateResourceRequest":{
				"required":[
					"my_string",
					"my_int"
				],
				"properties":{
					"my_string":{
						"description":"a string",
						"type":"string"
					},
					"my_int":{
						"description":"an int",
						"type":"integer",
						"format":"int64",
						"maximum":100,
						"minimum":0
					}
				}
			},
			"ResourceResponse":{
				"required":[
					"output_string",
					"output_int"
				],
				"properties":{
					"output_string":{
						"description":"a string",
						"type":"string"
					},
					"output_int":{
						"description":"an int",
						"type":"integer",
						"format":"int64",
						"maximum":100,
						"minimum":0
					}
				}
			},
			"ErrorResponse":{
				"required":[
					"code",
					"internal_code",
					"message"
				],
				"properties":{
					"code":{
						"type":"integer",
						"format":"int64"
					},
					"internal_code":{
						"type":"string"
					},
					"message":{
						"type":"string"
					}
				}
			}
		}
	}
	`)
}
