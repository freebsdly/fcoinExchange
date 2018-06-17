package fcoin

import (
	"fmt"
	"sort"
	"strings"
)

type KV struct {
	Key   string
	Value string
}

type KVSlice []KV

func (s KVSlice) Len() int           { return len(s) }
func (s KVSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s KVSlice) Less(i, j int) bool { return s[i].Key < s[j].Key }

func SortMap(m map[string]string, sep string) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	kvs := make(KVSlice, 0)
	for k, v := range m {
		kvs = append(kvs, KV{Key: k, Value: v})
	}

	sort.Sort(kvs)

	var s = make([]string, 0)
	for _, v := range kvs {
		s = append(s, fmt.Sprintf("%s=%v", v.Key, v.Value))
	}

	return strings.Join(s, sep)
}
