package utils

import (
	"encoding/json"
	"gorm.io/datatypes"
)

// ConvertJSONToMap converts datatypes.JSON to map[string]interface{}
func ConvertJSONToMap(jsonData datatypes.JSON) map[string]interface{} {
	if len(jsonData) == 0 {
		return make(map[string]interface{})
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		// Return empty map if unmarshal fails
		return make(map[string]interface{})
	}
	
	return result
}

// ConvertMapToJSON converts map[string]interface{} to datatypes.JSON
func ConvertMapToJSON(data map[string]interface{}) datatypes.JSON {
	if data == nil {
		data = make(map[string]interface{})
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Return empty JSON object if marshal fails
		return datatypes.JSON("{}")
	}
	
	return datatypes.JSON(jsonData)
}