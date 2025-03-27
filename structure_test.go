package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const sampleJson string = `
{
"data": [
    {
      "id": "stock1",
      "some_key": "1",
      "quotes": [{
          "currency": "USD",
          "price": 87
      }]
    },
    {
      "id": "stock14",
      "some_key": "14",
      "quotes": [{
          "currency": "USD",
          "price": 87
      }]
    }
  ]
}
`

type payload struct {
	Data []struct {
		ID string `json:"id"`
		// Added because I was not sure what the use of the key literal was.
		SomeKey string `json:"some_key"`
		Quote   []struct {
			Currency string  `json:"currency"`
			Price    float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

const stockQuotesSchema string = `
{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "$id": "https://example.com/stockquotes.schema.json",
    "title": "Stock Quotes",
    "description": "Quotes for stocks",
    "type": "object",
    "examples": [
        {
            "data": [
                {
                    "id": "stock1",
                    "some_key": "1",
                    "quote": [
                        {
                            "currency": "USD",
                            "price": 87
                        }
                    ]
                },
                {
                    "id": "stock14",
                    "some_key": "14",
                    "quote": [
                        {
                            "currency": "USD",
                            "price": 87
                        }
                    ]
                }
            ]
        }
    ],
    "$defs": {
        "quote": {
            "type": "object",
            "description": "A quote for a stock in a specific currency",
            "title": "Quote",
            "examples": [
                {
                    "currency": "USD",
                    "price": 87
                }
            ],
            "properties": {
                "currency": {
                    "type": "string"
                },
                "price": {
                    "type": [
                        "number",
                        "null"
                    ]
                }
            },
            "required": [
                "currency",
                "price"
            ]
        }
    },
    "properties": {
        "data": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "id": {
                        "type": "string",
                        "description": "Unique identifier for the stock",
                        "examples": [
                            "stock1",
                            "stock14"
                        ]
                    },
                    "some_key": {
                        "type": "string"
                    },
                    "quotes": {
                        "type": "array",
                        "description": "List of quotes for the stock in different currencies",
                        "title": "Quotes",
                        "examples": [
                            [
                                {
                                    "currency": "USD",
                                    "price": 87
                                },
                                {
                                    "currency": "EUR",
                                    "price": 75
                                }
                            ]
                        ],
                        "items": {
                            "$ref": "#/$defs/quote"
                        },
                        "minItems": 1
                    }
                },
                "required": [
                    "id",
                    "some_key",
                    "quotes"
                ]
            }
        }
    },
    "required": [
        "data"
    ]
}
`

func TestSchema(t *testing.T) {
	// Unmarshal the JSON schema
	stockSchema, err := jsonschema.UnmarshalJSON(strings.NewReader(stockQuotesSchema))
	if err != nil {
		t.Fatalf("Error unmarshalling schema: %v", err)
	}

	// Unmarshal the sample JSON data. Note that 'stockData' is an actual map[string]any
	// that contains the data from the JSON. This is a generic type that can hold arbitrary JSON data
	// which we can not only use for validation, but also to access the data, although it is not strongly typed
	// like the 'payload' struct.
	stockData, err := jsonschema.UnmarshalJSON(strings.NewReader(sampleJson))
	if err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// We create a new compiler and add the schema to it.
	c := jsonschema.NewCompiler()
	// Note that "stockquotes.schema.json" is just a placeholder name and is not a real file.
	if err := c.AddResource("stockquotes.schema.json", stockSchema); err != nil {
		t.Fatalf("Error adding resource: %v", err)
	}
	// Compile the schema. This will check for any errors in the schema.
	// The schema is compiled and stored in the compiler.
	sch, err := c.Compile("stockquotes.schema.json")
	if err != nil {
		t.Fatalf("Error compiling schema: %v", err)
	}

	// Validate the JSON data against the schema.
	err = sch.Validate(stockData)
	if err != nil {
		t.Errorf("Error validating JSON: %v", err)
	}

}

func TestIncompleteUnmarshal(t *testing.T) {
	p := new(payload)
	err := json.Unmarshal([]byte(sampleJson), p)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
	}
	if len(p.Data) != 2 {
		t.Errorf("Expected 2 items in data, got %d", len(p.Data))
	}
	if p.Data[0].ID != "stock1" {
		t.Errorf("Expected ID to be 'stock1', got '%s'", p.Data[0].ID)
	}
	if p.Data[0].SomeKey != "1" {
		t.Errorf("Expected SomeKey to be '1', got '%s'", p.Data[0].SomeKey)
	}
	if p.Data[1].ID != "stock14" {
		t.Errorf("Expected ID to be 'stock14', got '%s'", p.Data[1].ID)
	}
	if p.Data[1].SomeKey != "14" {
		t.Errorf("Expected SomeKey to be '14', got '%s'", p.Data[1].SomeKey)
	}

	defer func() {
		// This will catch the panic and allow us to check if it was caused by the missing field.
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	// This will cause a panic because the field 'quote' is not present in the struct
	// and the JSON will not be unmarshalled correctly.
	t.Logf("Quote: %v", p.Data[0].Quote[0].Currency)
}
