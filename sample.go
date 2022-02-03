package main

type Example struct {
	IntField       int `json:"field_1"`
	StringField    string
	Custom         []CustomStruct
	CustomEmbedded struct {
		First string
	} `json:"embedded"`
}

type CustomStruct struct {
	Truthy     bool
	ListString []string
}
