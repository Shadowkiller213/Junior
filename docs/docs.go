package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Ruslan Akuev",
            "email": "ruslanakuev312@gmail.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Returns service health status",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "Health check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/subscriptions": {
            "get": {
                "description": "Retrieve a list of subscriptions with optional filters",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subscriptions"
                ],
                "summary": "List subscriptions",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Filter by user UUID",
                        "name": "user_id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter by service name (case-insensitive)",
                        "name": "service_name",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/model.Subscription"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new subscription record",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subscriptions"
                ],
                "summary": "Create a subscription",
                "parameters": [
                    {
                        "description": "Subscription data",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.CreateRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/model.Subscription"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/subscriptions/total": {
            "get": {
                "description": "Calculate total cost of subscriptions over a period, with optional filters by user and service",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subscriptions"
                ],
                "summary": "Get total subscription cost",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Period start in MM-YYYY format",
                        "name": "start_period",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Period end in MM-YYYY format",
                        "name": "end_period",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Filter by user UUID",
                        "name": "user_id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter by service name (case-insensitive)",
                        "name": "service_name",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.TotalResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/subscriptions/{id}": {
            "get": {
                "description": "Retrieve a single subscription by its UUID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subscriptions"
                ],
                "summary": "Get a subscription by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subscription UUID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Subscription"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "put": {
                "description": "Fully update an existing subscription by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subscriptions"
                ],
                "summary": "Update a subscription",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subscription UUID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated subscription data",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.UpdateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Subscription"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete a subscription by ID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subscriptions"
                ],
                "summary": "Delete a subscription",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subscription UUID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.CreateRequest": {
            "type": "object",
            "properties": {
                "end_date": {
                    "type": "string"
                },
                "price": {
                    "type": "integer"
                },
                "service_name": {
                    "type": "string"
                },
                "start_date": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.Subscription": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "end_date": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "price": {
                    "type": "integer"
                },
                "service_name": {
                    "type": "string"
                },
                "start_date": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.TotalResponse": {
            "type": "object",
            "properties": {
                "total": {
                    "type": "integer"
                }
            }
        },
        "model.UpdateRequest": {
            "type": "object",
            "properties": {
                "end_date": {
                    "type": "string"
                },
                "price": {
                    "type": "integer"
                },
                "service_name": {
                    "type": "string"
                },
                "start_date": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Subscriptions API",
	Description:      "REST service for aggregating user online subscription data.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
