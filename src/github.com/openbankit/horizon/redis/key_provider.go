package redis

import "strings"

type namespace string

const (
	namespace_account_stats namespace = "as:"
	namespace_processed_op namespace = "pop:"
)

func getKey(ns namespace, keyParts... string) string {
	return string(ns) + strings.Join(keyParts, ":")
}
