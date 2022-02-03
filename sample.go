package main

type Example struct {
	Name           string
	Address        string
	Phone          string
	Email          string
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
