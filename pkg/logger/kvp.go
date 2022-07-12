package logger

// KVP is a Key-Value Pair. Also called an attribute-value pair.
// It's a data representation in computing systems and applications
type KVP struct {
	Key   string
	Value interface{}
}

// KeyValue returns the key and value of KVP
func (kv KVP) KeyValue() (string, interface{}) {
	return kv.Key, kv.Value
}
