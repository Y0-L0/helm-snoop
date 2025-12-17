package snooper

func flattenValues(prefix string, v interface{}, out map[string]struct{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenValues(key, child, out)
		}
	case map[interface{}]interface{}:
		for rk, child := range val {
			ks, ok := rk.(string)
			if !ok {
				continue
			}
			key := ks
			if prefix != "" {
				key = prefix + "." + ks
			}
			flattenValues(key, child, out)
		}
	case []interface{}:
		// arrays: mark container key only (leaf value indices are not keys)
		if prefix != "" {
			out[prefix] = struct{}{}
		}
		// do not descend into elements for key generation
	default:
		if prefix != "" {
			out[prefix] = struct{}{}
		}
	}
}
