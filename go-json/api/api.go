package api

type NamingStrategy func(flags uint16, key string) string
