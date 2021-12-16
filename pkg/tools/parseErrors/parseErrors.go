package parseErrors

func ErrorResponse(err error) map[string]interface{} {
	return map[string]interface{}{"error": err.Error()}
}
