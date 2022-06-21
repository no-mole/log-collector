package logger_center

import "strconv"

type Config struct {
	Outputs []*Output `json:"outputs"`
}

type Output struct {
	Type   string            `json:"type"`
	Tags   []string          `json:"tags"`
	Fields map[string]string `json:"fields"`
}

func (o *Output) GetString(key string, defaultVal string) string {
	if val, ok := o.Fields[key]; ok {
		return val
	}
	return defaultVal
}

func (o *Output) GetInt64(key string, defaultVal int64) int64 {
	if val, ok := o.Fields[key]; ok {
		intVal, _ := strconv.Atoi(val)
		return int64(intVal)
	}
	return defaultVal
}
