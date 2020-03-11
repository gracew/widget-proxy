package config

import "os"

const ParseURL = "http://parse:1337/parse/"

const CustomLogicUrl = "http://custom-logic:8080/"

const AuthPath = "/app/auth.json"

const CustomLogicPath = "/app/customLogic.json"

var (
	APIName = os.Getenv("API_NAME")
)
