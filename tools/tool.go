package tools

func ParsingMap(mapData map[string]interface{}, key string) (map[string]interface{}, bool) {
	if value, ok := mapData[key]; ok {
		data := value.(map[string]interface{})
		return data, true
	}
	return nil, false
}

func ParsingMapStr(mapData map[string]interface{}, key string) (string, bool) {
	if value, ok := mapData[key]; ok {
		data := value.(string)
		return data, true
	}
	return "", false
}

func ParsingMapSlice(mapData map[string]interface{}, key string) ([]interface{}, bool) {
	if value, ok := mapData[key]; ok {
		data := value.([]interface{})
		return data, true
	}
	return nil, false
}
