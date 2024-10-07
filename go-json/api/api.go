package api

type NamingStrategy func(flags uint16, key string) string

type QuoteNumberStrategy func(numBitSize uint8, negative, unsigned bool, u64 uint64) bool
